package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	database "songtor/config"
)

var db *gorm.DB
var sqsClient *sqs.Client
var validate *validator.Validate

func init() {
	var err error
	cfg := database.Config{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   os.Getenv("DB_NAME"),
	}

	db, err = database.ConnectGorm(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Create SQS Client
	awsCfg, _ := config.LoadDefaultConfig(context.Background())
	sqsClient = sqs.NewFromConfig(awsCfg)

	validate = validator.New()
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		return name
	})
}

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest(ctx context.Context, sqsEvent events.SQSEvent) error {
	for _, message := range sqsEvent.Records {
		err := processMessage(ctx, message)
		if err != nil {
			return err
		}
	}
	return nil
}

func processMessage(ctx context.Context, msg events.SQSMessage) error {
	attrID, ok1 := msg.MessageAttributes["message_id"]
	attrReply, ok2 := msg.MessageAttributes["reply_to"]
	if !ok1 || !ok2 {
		log.Printf("Missing required attributes: message_id or reply_to")
		return nil
	}

	msgID := aws.ToString(attrID.StringValue)
	replyTo := aws.ToString(attrReply.StringValue)

	var req DivertRequest
	if err := json.Unmarshal([]byte(msg.Body), &req); err != nil {
		log.Printf("Failed to unmarshal body: %v", err)
		return nil
	}

	var response DivertResponse
	// Business Logic & Database Transaction
	err := db.Transaction(func(tx *gorm.DB) error {
		// Idempotency Check
		processed := ProcessedMessage{MessageID: msgID}
		if err := tx.Create(&processed).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				log.Printf("Message %s already processed. Skipping.", msgID)
				return nil
			}
			return err
		}

		// Validation Rules
		if err := validate.Struct(req); err != nil {
			response = createRejectResponse(req, "VALIDATION_FAILED", formatMessageError(err))
			return nil
		}

		// Business Status Check
		var notification NotificationRecord
		if err := tx.Where("notification_id = ?", req.NotificationID).First(&notification).Error; err != nil {
			response = createRejectResponse(req, "NOT_FOUND", "Notification not found")
			return nil
		}

		if notification.Status != "EN_ROUTE" {
			response = createRejectResponse(req, "TOO_LATE", "Ambulance has already arrived")
			return nil
		}

		// Update Hospital
		if err := tx.Model(&notification).Update("hospital_id", req.NewHospitalID).Error; err != nil {
			return err
		}

		response = DivertResponse{
			NotificationID: req.NotificationID,
			RequestID:      req.RequestID,
			Status:         "CONFIRMED",
			NewHospitalID:  req.NewHospitalID,
		}
		return nil
	})

	// Manage Technical Errors (DB, SQS, etc.)
	if err != nil {
		return err // Let Lambda retry the message
	}

	// Send async response
	return publishResponse(ctx, replyTo, msgID, response)
}

func createRejectResponse(req DivertRequest, code, msg string) DivertResponse {
	return DivertResponse{
		NotificationID: req.NotificationID,
		RequestID:      req.RequestID,
		Status:         "REJECTED",
		ReasonCode:     code,
		ReasonMessage:  msg,
	}
}

func publishResponse(ctx context.Context, queueURL, correlationID string, resp DivertResponse) error {
	body, _ := json.Marshal(resp)

	msgType := "AmbulanceDivertConfirmed"
	if resp.Status == "REJECTED" {
		msgType = "AmbulanceDivertRejected"
	}

	_, err := sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"message_type":   {DataType: aws.String("String"), StringValue: aws.String(msgType)},
			"message_id":     {DataType: aws.String("String"), StringValue: aws.String(uuid.New().String())},
			"correlation_id": {DataType: aws.String("String"), StringValue: aws.String(correlationID)},
			"sent_at":        {DataType: aws.String("String"), StringValue: aws.String(time.Now().Format(time.RFC3339))},
		},
	})
	return err
}

func formatMessageError(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok {
		fe := ve[0]
		switch fe.Tag() {
		case "required":
			return fmt.Sprintf("%s required", fe.Field())
		case "nefield":
			return fmt.Sprintf("%s must be different from %s", fe.Field(), fe.Param())
		}
	}
	return "invalid request payload"
}
