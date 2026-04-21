package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type SQSAPI interface {
	SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error)
}

type App struct {
	db        *gorm.DB
	sqsClient SQSAPI
}

func (a *App) HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		messageID := message.MessageId

		var replyToQueueURL string
		if attr, ok := message.MessageAttributes["reply_to"]; ok {
			replyToQueueURL = *attr.StringValue
		} else {
			fmt.Printf("Warning: reply_to attribute is missing for message %s\n", messageID)
			continue
		}

		var req HospitalInboundLoadRequest
		if err := json.Unmarshal([]byte(message.Body), &req); err != nil {
			fmt.Printf("Failed to unmarshal body: %v\n", err)
			continue
		}

		if req.RequestID == "" || len(req.TargetHospitalID) == 0 || req.RequestedBy == "" {
			rejectResp := HospitalInboundLoadResponse{
				RequestID:     req.RequestID,
				Status:        "REJECTED",
				ReasonCode:    "INVALID_HOSPITAL_LIST",
				ReasonMessage: "request_id, target_hospital_id, and requested_by must not be empty.",
			}
			a.sendResponse(ctx, rejectResp, messageID, replyToQueueURL)
			continue
		}

		var results []DBResult
		err := a.db.Table("notification_records"). // Fixed table name to match models
								Joins("JOIN patient_data ON notification_records.notification_id = patient_data.notification_id").
								Select("hospital_id, triage_level, count(*) as count").
								Where("hospital_id IN ? AND status = ?", req.TargetHospitalID, "EN_ROUTE").
								Group("hospital_id, triage_level").
								Scan(&results).Error

		if err != nil {
			fmt.Printf("Database query error: %v\n", err)
			return err
		}

		loadMap := make(map[string]*InboundLoad)
		for _, hid := range req.TargetHospitalID {
			loadMap[hid] = &InboundLoad{
				HospitalID:    hid,
				TotalEnRoute:  0,
				TriageSummary: TriageSummary{},
			}
		}

		for _, row := range results {
			if load, exists := loadMap[row.HospitalID]; exists {
				load.TotalEnRoute += row.Count
				switch row.TriageLevel {
				case "RED":
					load.TriageSummary.Red = row.Count
				case "YELLOW":
					load.TriageSummary.Yellow = row.Count
				case "GREEN":
					load.TriageSummary.Green = row.Count
				}
			}
		}

		var inboundLoads []InboundLoad
		for _, load := range loadMap {
			inboundLoads = append(inboundLoads, *load)
		}

		successResp := HospitalInboundLoadResponse{
			RequestID:    req.RequestID,
			InboundLoads: inboundLoads,
			Status:       "CONFIRMED",
		}

		if err := a.sendResponse(ctx, successResp, messageID, replyToQueueURL); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) sendResponse(ctx context.Context, payload HospitalInboundLoadResponse, correlationID string, replyToQueueURL string) error {
	responseJSON, _ := json.Marshal(payload)
	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(replyToQueueURL),
		MessageBody: aws.String(string(responseJSON)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"message_type":   {DataType: aws.String("String"), StringValue: aws.String("HospitalInboundLoadProvided")},
			"correlation_id": {DataType: aws.String("String"), StringValue: aws.String(correlationID)},
			"message_id":     {DataType: aws.String("String"), StringValue: aws.String(uuid.New().String())},
			"sent_at":        {DataType: aws.String("String"), StringValue: aws.String(time.Now().UTC().Format(time.RFC3339))},
		},
	}

	fmt.Printf("Sending response to %s: %s\n", replyToQueueURL, string(responseJSON))
	_, err := a.sqsClient.SendMessage(ctx, input)
	return err
}

type HospitalInboundLoadRequest struct {
	RequestID        string   `json:"request_id"`
	TargetHospitalID []string `json:"target_hospital_id"`
	RequestedBy      string   `json:"requested_by"`
}

type TriageSummary struct {
	Red    int `json:"RED"`
	Yellow int `json:"YELLOW"`
	Green  int `json:"GREEN"`
}

type InboundLoad struct {
	HospitalID    string        `json:"hospital_id"`
	TotalEnRoute  int           `json:"total_en_route"`
	TriageSummary TriageSummary `json:"triage_summary"`
}

type HospitalInboundLoadResponse struct {
	RequestID     string        `json:"request_id"`
	InboundLoads  []InboundLoad `json:"inbound_loads,omitempty"`
	Status        string        `json:"status"`
	ReasonCode    string        `json:"reason_code,omitempty"`
	ReasonMessage string        `json:"reason_message,omitempty"`
}

type DBResult struct {
	HospitalID  string
	TriageLevel string
	Count       int
}

func main() {
	dsn := os.Getenv("DB_DSN")
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(err)
	}

	cfg, _ := config.LoadDefaultConfig(context.TODO())
	app := &App{
		db:        db,
		sqsClient: sqs.NewFromConfig(cfg),
	}

	lambda.Start(app.HandleRequest)
}
