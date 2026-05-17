package handlers

import (
	"encoding/json"
	"fmt"
	"songtor/internal/dto"
	"songtor/internal/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type NotificationHandler struct {
	db                             *gorm.DB
	PatientReportedTopicArn        string
	CriticalCaseTopicArn           string
	HOSPITAL_RESOURCE_SERVICE_HOST string
}

func NewNotificationHandler(db *gorm.DB, PatientReportedTopicArn string, CriticalCaseTopicArn string, HOSPITAL_RESOURCE_SERVICE_HOST string) *NotificationHandler {
	return &NotificationHandler{
		db:                             db,
		PatientReportedTopicArn:        PatientReportedTopicArn,
		CriticalCaseTopicArn:           CriticalCaseTopicArn,
		HOSPITAL_RESOURCE_SERVICE_HOST: HOSPITAL_RESOURCE_SERVICE_HOST,
	}
}

func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	// Parse the request body
	var req dto.EmergencyRequest
	if err := c.BodyParser(&req); err != nil {
		fmt.Printf("Failed to parse request body: %v\n", err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request"})
	}

	// Validate the request
	if err := ValidateEmergencyRequest(&req, c); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
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

	if h.HOSPITAL_RESOURCE_SERVICE_HOST != "" {
		isResourceAvailable := false

		// Check hospital resources
		agent := fiber.Get(h.HOSPITAL_RESOURCE_SERVICE_HOST + "/v1/hospitals/" + req.HospitalID + "/resources")
		agent.Timeout(20 * time.Second)

		statusCode, body, errs := agent.Bytes()
		var resource dto.HospitalResourceResponse
		if errs != nil || statusCode != fiber.StatusOK {
			fmt.Printf("Failed to fetch hospital resources: %v\n", body)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to fetch hospital resources"})
		} else {
			// Parse the hospital resource response
			if err := json.Unmarshal(body, &resource); err != nil {
				fmt.Printf("Failed to parse hospital resource response: %v\n", err)
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to parse JSON response"})
			} else {
				isResourceAvailable = true
			}
		}

		if isResourceAvailable {
			// Check if any resource is in CRITICAL status
			for _, resource := range resource.Data.Resources {
				if resource.ResourceStatus == "CRITICAL" {
					fmt.Printf("Hospital resources are in CRITICAL status: %v\n", resource)
					return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "Hospital resources are currently unavailable"})
				}
			}
		}
	} else {
		fmt.Println("HOSPITAL_RESOURCE_SERVICE_HOST is not configured, skipping hospital resource check")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create notification"})
	}

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
		fmt.Printf("Failed to create notification record: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create notification"})
	}

	// Create outbox event for PatientReportedSNS if topic ARN is configured
	if h.PatientReportedTopicArn != "" {
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
			fmt.Printf("Failed to marshal PatientReportedSNS message: %v\n", err)
		}

		outboxEvent := models.OutboxEvent{
			TopicARN: h.PatientReportedTopicArn,
			Payload:  string(messageBytes),
			Status:   "PENDING",
		}

		if err := tx.Create(&outboxEvent).Error; err != nil {
			tx.Rollback()
			fmt.Printf("Failed to create PatientReportedSNS outbox event: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create PatientReported outbox event"})
		}
	}

	// Create outbox event for CriticalCaseSNS if topic ARN is configured
	if h.CriticalCaseTopicArn != "" {
		criticalCaseReq := dto.CriticalCaseNotification{
			HospitalID:  req.HospitalID,
			TriageLevel: req.TriageLevel,
			Status:      "EN_ROUTE",
		}
		messageBytes, err := json.Marshal(criticalCaseReq)
		if err != nil {
			fmt.Printf("Failed to marshal CriticalCaseSNS message: %v\n", err)
		}

		outboxEvent := models.OutboxEvent{
			TopicARN: h.CriticalCaseTopicArn,
			Payload:  string(messageBytes),
			Status:   "PENDING",
		}

		if err := tx.Create(&outboxEvent).Error; err != nil {
			tx.Rollback()
			fmt.Printf("Failed to create CriticalCase outbox event: %v\n", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create CriticalCase outbox event"})
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
			"error": "Missing required fields",
			"code":  "VALIDATION_ERROR",
		})
	}

	if req.TriageLevel != "RED" && req.TriageLevel != "YELLOW" && req.TriageLevel != "GREEN" && req.TriageLevel != "BLACK" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid triage level",
			"code":  "VALIDATION_ERROR",
		})
	}

	if req.PatientInfo.AgeCategory != "ADULT" && req.PatientInfo.AgeCategory != "INFANTS" && req.PatientInfo.AgeCategory != "TEEN" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid age category",
			"code":  "VALIDATION_ERROR",
		})
	}

	if req.PatientInfo.LifeStatus != "ALIVE" && req.PatientInfo.LifeStatus != "DECEASED" && req.PatientInfo.LifeStatus != "UNKNOWN" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid life status",
			"code":  "VALIDATION_ERROR",
		})
	}

	return nil
}
