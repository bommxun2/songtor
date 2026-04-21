package handlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"songtor/internal/dto"
	"songtor/internal/models"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to connect database: %v", err)
	}

	err = db.AutoMigrate(&models.NotificationRecord{}, &models.PatientData{}, &models.OutboxEvent{})
	if err != nil {
		t.Fatalf("failed to migrate database: %v", err)
	}

	return db
}

func TestCreateNotification(t *testing.T) {
	db := setupTestDB(t)
	app := fiber.New()
	h := NewNotificationHandler(db, "arn:aws:sns:topic", "")

	app.Post("/notifications", h.CreateNotification)

	t.Run("Success", func(t *testing.T) {
		reqBody := dto.EmergencyRequest{
			HospitalID:  "hosp1",
			AmbulanceID: "amb1",
			PatientInfo: dto.PatientInfo{
				Gender:       "MALE",
				AgeCategory:  "ADULT",
				PhysicalDesc: "Tattoo on arm",
				LifeStatus:   "ALIVE",
			},
			TriageLevel:    "RED",
			Symptom:        "Chest pain",
			AttachmentURLs: []string{"http://image.com/1"},
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/notifications", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

		var res map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&res)
		assert.NotEmpty(t, res["notification_id"])

		// Verify database
		var count int64
		db.Model(&models.NotificationRecord{}).Count(&count)
		assert.Equal(t, int64(1), count)

		var outboxCount int64
		db.Model(&models.OutboxEvent{}).Count(&outboxCount)
		assert.Equal(t, int64(1), outboxCount)
	})

	t.Run("Validation Error", func(t *testing.T) {
		reqBody := dto.EmergencyRequest{
			HospitalID: "", // Missing field
		}

		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/notifications", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	})
}
