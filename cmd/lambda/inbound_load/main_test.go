package main

import (
	"context"
	"encoding/json"
	"songtor/internal/models"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockSQS struct {
	mock.Mock
}

func (m *MockSQS) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(*sqs.SendMessageOutput), args.Error(1)
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	// AutoMigrate tables used in join query
	err = db.AutoMigrate(&models.NotificationRecord{}, &models.PatientData{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestHandleRequest(t *testing.T) {
	db := setupTestDB(t)
	mockSQS := new(MockSQS)
	app := &App{db: db, sqsClient: mockSQS}

	// Seed data
	db.Create(&models.NotificationRecord{
		ID:         "notif1",
		HospitalID: "hosp1",
		Status:     "EN_ROUTE",
		PatientData: models.PatientData{
			TriageLevel: "RED",
			Symptom:     "Heart attack",
		},
	})

	t.Run("Success - Single Hospital", func(t *testing.T) {
		reqBody, _ := json.Marshal(HospitalInboundLoadRequest{
			RequestID:        "req1",
			TargetHospitalID: []string{"hosp1"},
			RequestedBy:      "user1",
		})

		event := events.SQSEvent{
			Records: []events.SQSMessage{
				{
					MessageId: "msg1",
					Body:      string(reqBody),
					MessageAttributes: map[string]events.SQSMessageAttribute{
						"reply_to": {StringValue: aws.String("http://reply-queue")},
					},
				},
			},
		}

		mockSQS.On("SendMessage", mock.Anything, mock.MatchedBy(func(input *sqs.SendMessageInput) bool {
			var resp HospitalInboundLoadResponse
			json.Unmarshal([]byte(*input.MessageBody), &resp)
			return resp.Status == "CONFIRMED" && len(resp.InboundLoads) == 1 && resp.InboundLoads[0].TriageSummary.Red == 1
		})).Return(&sqs.SendMessageOutput{}, nil)

		err := app.HandleRequest(context.TODO(), event)
		assert.NoError(t, err)
		mockSQS.AssertExpectations(t)
	})

	t.Run("Rejected - Missing Fields", func(t *testing.T) {
		reqBody, _ := json.Marshal(HospitalInboundLoadRequest{
			RequestID: "req2",
			// Missing TargetHospitalID
		})

		event := events.SQSEvent{
			Records: []events.SQSMessage{
				{
					MessageId: "msg2",
					Body:      string(reqBody),
					MessageAttributes: map[string]events.SQSMessageAttribute{
						"reply_to": {StringValue: aws.String("http://reply-queue")},
					},
				},
			},
		}

		mockSQS.On("SendMessage", mock.Anything, mock.MatchedBy(func(input *sqs.SendMessageInput) bool {
			var resp HospitalInboundLoadResponse
			json.Unmarshal([]byte(*input.MessageBody), &resp)
			return resp.Status == "REJECTED"
		})).Return(&sqs.SendMessageOutput{}, nil)

		err := app.HandleRequest(context.TODO(), event)
		assert.NoError(t, err)
		mockSQS.AssertExpectations(t)
	})
}
