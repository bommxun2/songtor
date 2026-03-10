package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	database "songtor/config"
	"songtor/internal/handlers"
	"songtor/internal/models"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found. Falling back to system environment variables.")
	}

	// Load database configulation
	dbConfig := database.Config{
		User:     getEnv("DB_USER", "test_user"),
		Password: getEnv("DB_PASSWORD", "test_password"),
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnv("DB_PORT", "3306"),
		DBName:   getEnv("DB_NAME", "test_db"),
	}

	// Connect to MySQL database
	db, err := database.ConnectGorm(dbConfig)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err)
	}

	// Run database migrations
	log.Println("Running Auto Migration...")
	err = db.AutoMigrate(&models.NotificationRecord{}, &models.PatientData{}, &models.ProcessedMessage{})
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	// Initialize Fiber App
	app := fiber.New()

	notificationHandler := handlers.NewNotificationHandler(db)

	v1 := app.Group("/v1")
	v1.Post("/notifications", notificationHandler.CreateNotification)

	log.Println("Server is running on port 8080")
	if err := app.Listen(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}
