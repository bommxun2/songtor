package workers

import (
	"context"
	"errors"
	"songtor/internal/models"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// MockSNS is a mock type for the SNSAPI interface
type MockSNS struct {
	mock.Mock
}

func (m *MockSNS) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sns.PublishOutput), args.Error(1)
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&models.OutboxEvent{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestProcessOutboxEvents(t *testing.T) {
	t.Run("Successfully publish and delete event", func(t *testing.T) {
		db := setupTestDB(t)
		mockSNS := new(MockSNS)

		event := models.OutboxEvent{
			TopicARN: "arn:test",
			Payload:  "test payload",
			Status:   "PENDING",
		}
		db.Create(&event)

		mockSNS.On("Publish", mock.Anything, mock.MatchedBy(func(input *sns.PublishInput) bool {
			return *input.TopicArn == "arn:test" && *input.Message == "test payload"
		})).Return(&sns.PublishOutput{}, nil)

		processOutboxEvents(db, mockSNS)

		// Verify event was deleted
		var count int64
		db.Model(&models.OutboxEvent{}).Count(&count)
		assert.Equal(t, int64(0), count)
		mockSNS.AssertExpectations(t)
	})

	t.Run("Handle SNS failure", func(t *testing.T) {
		db := setupTestDB(t)
		mockSNS := new(MockSNS)

		event := models.OutboxEvent{
			TopicARN: "arn:test",
			Payload:  "test payload",
			Status:   "PENDING",
		}
		db.Create(&event)

		mockSNS.On("Publish", mock.Anything, mock.Anything).Return(nil, errors.New("sns error"))

		processOutboxEvents(db, mockSNS)

		// Verify status updated to FAILED
		var updatedEvent models.OutboxEvent
		db.First(&updatedEvent, event.ID)
		assert.Equal(t, "FAILED", updatedEvent.Status)
		mockSNS.AssertExpectations(t)
	})
}
