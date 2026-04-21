package workers

import (
	"context"
	"fmt"
	"time"

	"songtor/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"gorm.io/gorm"
)

type SNSAPI interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

func StartOutboxWorker(db *gorm.DB, snsClient SNSAPI) {
	go func() {
		for {
			processOutboxEvents(db, snsClient)
			time.Sleep(5 * time.Second)
		}
	}()
}

func processOutboxEvents(db *gorm.DB, snsClient SNSAPI) {
	var events []models.OutboxEvent

	// Get pending events
	if err := db.Where("status = ?", "PENDING").Limit(50).Find(&events).Error; err != nil {
		fmt.Printf("[Worker Error] Failed to fetch outbox events: %v\n", err)
		return
	}

	for _, event := range events {
		// Update status to PROCESSING to avoid duplicate processing
		db.Model(&event).Update("status", "PROCESSING")

		publishInput := &sns.PublishInput{
			TopicArn: aws.String(event.TopicARN),
			Message:  aws.String(event.Payload),
		}

		_, err := snsClient.Publish(context.TODO(), publishInput)
		if err != nil {
			fmt.Printf("[Worker Error] Failed to publish event ID %d: %v\n", event.ID, err)
			db.Model(&event).Update("status", "FAILED")
		} else {
			fmt.Printf("[Worker Success] Event ID %d published to SNS\n", event.ID)
			db.Delete(&event)
		}
	}
}
