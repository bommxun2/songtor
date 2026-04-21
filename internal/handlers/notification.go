package handlers

import (
	"encoding/json"
	"fmt"
	"songtor/internal/dto"
	"songtor/internal/models"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	db                             *gorm.DB
	topicArn                       string
	HOSPITAL_RESOURCE_SERVICE_HOST string
}

func NewNotificationHandler(db *gorm.DB, topicArn string, HOSPITAL_RESOURCE_SERVICE_HOST string) *NotificationHandler {
	return &NotificationHandler{
		db:                             db,
		topicArn:                       topicArn,
		HOSPITAL_RESOURCE_SERVICE_HOST: HOSPITAL_RESOURCE_SERVICE_HOST,
	}
}

func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	// Parse the request body
	var req dto.EmergencyRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate the request
	if err := ValidateEmergencyRequest(&req, c); err != nil {
		return err
	}

	// Generate a unique notification ID
	notificationID := uuid.New().String()
	indempotencyKey := uuid.New()

	// Create the notification record
	notification := models.NotificationRecord{
		ID:             notificationID,
		HospitalID:     req.HospitalID,
		AmbulanceID:    req.AmbulanceID,
		IdempotencyKey: indempotencyKey.String(),
		PatientData: models.PatientData{
			TriageLevel:    req.TriageLevel,
			Symptom:        req.Symptom,
			CategoryAge:    &req.PatientInfo.AgeCategory,
			Gender:         &req.PatientInfo.Gender,
			Bp:             &req.Vitals.BP,
			Hr:             &req.Vitals.HR,
			Spo2:           &req.Vitals.SpO2,
			AttachmentURLs: req.AttachmentURLs,
		},
	}

	/*
		if h.HOSPITAL_RESOURCE_SERVICE_HOST != "" {
			isResourceAvailable := false

			// Check hospital resources
			agent := fiber.Get("http://" + h.HOSPITAL_RESOURCE_SERVICE_HOST + "/hospital/" + req.HospitalID + "/resources").
				agent.Timeout(20 * time.Second)

			statusCode, body, errs := agent.Bytes()
			if errs != nil || statusCode != fiber.StatusOK {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch hospital resources"})
			} else {
				// Parse the hospital resource response
				var resource HospitalResourceResponse
				if err := json.Unmarshal(body, &resource); err != nil {
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse JSON response"})
				} else {
					isResourceAvailable = true
				}
			}

			if isResourceAvailable {
				// Check if any resource is in CRITICAL status
				for _, resource := range resource.Resources {
					if resource.resourceStatus == "CRITICAL" {
						return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Hospital resources are currently unavailable"})
					}
				}
			}
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create notification"})
		}
	*/

	tx := h.db.Begin()
	defer func() {
		// If there is a panic, rollback the transaction to prevent partial commits
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create the notification record
	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create notification"})
	}

	// Create outbox event for SNS if topic ARN is configured
	if h.topicArn != "" {
		patientTransferReq := dto.PatientTransferRequest{
			DestinationHospitalID: req.HospitalID,
			Characteristics:       dto.Characteristics(req.PatientInfo),
			Media: dto.Media{
				MissingPersonPhoto: req.AttachmentURLs[0],
			},
			ReportedBy: "PreArrivalNotificationService",
		}
		messageBytes, err := json.Marshal(patientTransferReq)
		if err != nil {
			fmt.Printf("Failed to marshal SNS message: %v\n", err)
		}

		outboxEvent := models.OutboxEvent{
			TopicARN: h.topicArn,
			Payload:  string(messageBytes),
			Status:   "PENDING",
		}

		if err := tx.Create(&outboxEvent).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create outbox event"})
		}
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Transaction commit failed"})
	}

	// Return the response
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"notification_id": notification.ID,
		"status":          notification.Status,
		"message":         "Created notification successfully",
	})
}

func ValidateEmergencyRequest(req *dto.EmergencyRequest, c *fiber.Ctx) error {
	if req.HospitalID == "" || req.PatientInfo.PhysicalDesc == "" || req.PatientInfo.AgeCategory == "" ||
		req.PatientInfo.LifeStatus == "" || req.PatientInfo.Gender == "" || req.TriageLevel == "" || req.Symptom == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Missing required fields",
			"code":     "VALIDATION_ERROR",
			"trace_id": c.Get("X-Trace-ID"),
		})
	}

	if req.TriageLevel != "RED" && req.TriageLevel != "YELLOW" && req.TriageLevel != "GREEN" && req.TriageLevel != "BLACK" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid triage level",
			"code":     "VALIDATION_ERROR",
			"trace_id": c.Get("X-Trace-ID"),
		})
	}

	if req.PatientInfo.AgeCategory != "ADULT" && req.PatientInfo.AgeCategory != "INFANTS" && req.PatientInfo.AgeCategory != "TEEN" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid age category",
			"code":     "VALIDATION_ERROR",
			"trace_id": c.Get("X-Trace-ID"),
		})
	}

	if req.PatientInfo.LifeStatus != "ALIVE" && req.PatientInfo.LifeStatus != "DECEASED" && req.PatientInfo.LifeStatus != "UNKNOWN" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":    "Invalid life status",
			"code":     "VALIDATION_ERROR",
			"trace_id": c.Get("X-Trace-ID"),
		})
	}

	return nil
}
