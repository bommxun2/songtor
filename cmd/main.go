package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"gorm.io/gorm"

	database "songtor/config"
	"songtor/internal/handlers"
	"songtor/internal/models"
	workers "songtor/internal/worker"
)

func main() {
	_ = godotenv.Load()

	// 1. Database Setup with Retry
	var db *gorm.DB
	var err error
	dbConfig := database.Config{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   os.Getenv("DB_NAME"),
	}

	for i := 0; i < 10; i++ {
		db, err = database.ConnectGorm(dbConfig)
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database (attempt %d/10): %v", i+1, err)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		log.Fatalf("Database connection failed after retries: %v", err)
	}

	log.Println("Running migrations...")
	if err := db.AutoMigrate(&models.NotificationRecord{}, &models.PatientData{}, &models.OutboxEvent{}); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// 2. AWS Setup
	awsCfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Printf("Warning: Failed to load AWS config: %v", err)
	}
	snsClient := sns.NewFromConfig(awsCfg)

	// 3. Handlers & Routes
	app := fiber.New()
	notificationHandler := handlers.NewNotificationHandler(db, os.Getenv("SNS_TOPIC_ARN"), os.Getenv("HOSPITAL_RESOURCE_SERVICE_HOST"))
	listIncomingHandler := handlers.NewListIncomingHandler(db)

	v1 := app.Group("/v1")
	v1.Post("/notifications", notificationHandler.CreateNotification)
	v1.Get("/hospitals/:hospital_id/incoming-notifications", listIncomingHandler.ListIncomingPatients)

	// 4. Background Workers
	workers.StartOutboxWorker(db, snsClient)

	log.Println("Server starting on :8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
