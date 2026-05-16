package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"songtor/internal/dto"
	"songtor/internal/models"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestListNotifications(t *testing.T) {
	db := setupTestDB(t)
	app := fiber.New()
	h := NewListNotificationHandler(db)

	app.Get("/v1/hospitals/:hospital_id/incoming-notifications", h.ListNotifications)

	// Seed data
	ageStr := "45"
	genderStr := "M"
	bpStr := "80/50"
	hrInt := 120
	spo2Int := 85

	db.Create(&models.NotificationRecord{
		ID:             "NOTIF-20261027-005",
		HospitalID:     "HOS-001",
		AmbulanceID:    "AMB-BKK-009",
		Status:         "EN_ROUTE",
		IdempotencyKey: "test-key-1",
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		PatientData: models.PatientData{
			TriageLevel:    "RED",
			Symptom:        "Cardiac Arrest",
			CategoryAge:    &ageStr,
			Gender:         &genderStr,
			Bp:             &bpStr,
			Hr:             &hrInt,
			Spo2:           &spo2Int,
			AttachmentURLs: []string{"https://example.png"},
		},
	})

	t.Run("Success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/hospitals/HOS-001/incoming-notifications", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer mock-token")

		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var res dto.ListNotificationResponse
		json.NewDecoder(resp.Body).Decode(&res)

		assert.Equal(t, "HOS-001", res.HospitalID)
		assert.Len(t, res.Items, 1)

		item := res.Items[0]
		assert.Equal(t, "NOTIF-20261027-005", item.NotificationID)
		assert.Equal(t, "AMB-BKK-009", item.AmbulanceID)
		assert.Equal(t, 45, item.PatientInfo.Age)
		assert.Equal(t, "M", item.PatientInfo.Gender)
		assert.Equal(t, "RED", item.TriageLevel)
		assert.Equal(t, "Cardiac Arrest", item.Symptom)
		assert.Equal(t, "80/50", item.Vitals.BP)
		assert.Equal(t, 120, item.Vitals.HR)
		assert.Equal(t, 85, item.Vitals.SpO2)
		assert.Equal(t, "EN_ROUTE", item.Status)
		assert.Len(t, item.AttachmentURLs, 1)
		assert.Equal(t, "https://example.png", item.AttachmentURLs[0])
	})

	t.Run("Empty List for another hospital", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/v1/hospitals/HOS-002/incoming-notifications", nil)
		req.Header.Set("Content-Type", "application/json")

		resp, _ := app.Test(req)

		assert.Equal(t, fiber.StatusOK, resp.StatusCode)

		var res dto.ListNotificationResponse
		json.NewDecoder(resp.Body).Decode(&res)

		assert.Equal(t, "HOS-002", res.HospitalID)
		assert.Len(t, res.Items, 0)
	})
}
