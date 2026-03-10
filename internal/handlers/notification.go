package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"songtor/internal/models"
)

type NotificationHandler struct {
	Validator *validator.Validate
	DB        *gorm.DB
}

func NewNotificationHandler(db *gorm.DB) *NotificationHandler {
	v := validator.New()

	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
		return name
	})

	return &NotificationHandler{Validator: v, DB: db}
}

func (h *NotificationHandler) CreateNotification(c *fiber.Ctx) error {
	// Create a unique Trace ID for this request to help with debugging and log correlation
	traceID := uuid.New().String()

	// Pull out Idempotency-Key from Header
	idempKey := c.Get("Idempotency-Key")
	if idempKey == "" {
		return sendErrorResponse(c, fiber.StatusBadRequest, "Idempotency-Key header is required", traceID)
	}

	// Parse JSON Body
	req := new(models.CreateNotificationRequest)
	if err := c.BodyParser(req); err != nil {
		return sendErrorResponse(c, fiber.StatusBadRequest, "Invalid JSON body", traceID)
	}

	// Validate Request body
	if err := h.Validator.Struct(req); err != nil {
		return sendErrorResponse(c, fiber.StatusBadRequest, formatValidationError(err), traceID)
	}

	// Convert optional fields to pointers
	var agePtr *int
	if req.PatientInfo.Age > 0 {
		agePtr = &req.PatientInfo.Age
	}

	var genderPtr *string
	if req.PatientInfo.Gender != "" {
		genderPtr = &req.PatientInfo.Gender
	}

	var bpPtr *string
	if req.Vitals.Bp != "" {
		bpPtr = &req.Vitals.Bp
	}

	var hrPtr *int
	if req.Vitals.Hr > 0 {
		hrPtr = &req.Vitals.Hr
	}

	var spo2Ptr *int
	if req.Vitals.Spo2 > 0 {
		spo2Ptr = &req.Vitals.Spo2
	}

	// Convert Attachment URLs to JSON String
	attachBytes, _ := json.Marshal(req.AttachmentURLs)

	// Create Notification ID (e.g. NOTIF-20240601-1a2b)
	dateStr := time.Now().Format("20060102")
	notificationID := fmt.Sprintf("NOTIF-%s-%s", dateStr, uuid.New().String()[:4])

	// Map request data to NotificationRecord model
	record := models.NotificationRecord{
		ID:             notificationID,
		HospitalID:     req.HospitalID,
		AmbulanceID:    req.AmbulanceID,
		Status:         "EN_ROUTE",
		IdempotencyKey: idempKey,
		PatientData: models.PatientData{
			NotificationID: notificationID,
			TriageLevel:    req.TriageLevel,
			Symptom:        req.Symptom,
			PatientAge:     agePtr,
			PatientGender:  genderPtr,
			Bp:             bpPtr,
			Hr:             hrPtr,
			Spo2:           spo2Ptr,
			AttachmentURLs: string(attachBytes),
		},
	}

	// Create record in database
	if err := h.DB.Create(&record).Error; err != nil {
		// Check if error is due to duplicate Idempotency-Key
		if strings.Contains(err.Error(), "Duplicate entry") {
			return sendErrorResponse(c, fiber.StatusConflict, "Request already processed (Duplicate Idempotency-Key)", traceID)
		}

		// Case when database is unreachable or other internal error
		return sendErrorResponse(c, fiber.StatusInternalServerError, "Failed to save notification to database", traceID)
	}

	// Response with success
	return c.Status(fiber.StatusCreated).JSON(models.SuccessResponse{
		NotificationID: notificationID,
		Status:         "EN_ROUTE",
		Message:        "Notification received. ER team alerted.",
	})
}

func formatValidationError(err error) string {
	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		firstErr := validationErrors[0]
		switch firstErr.Tag() {
		case "required":
			return fmt.Sprintf("%s required", firstErr.Field())
		case "oneof":
			return fmt.Sprintf("%s must be one of RED, YELLOW, GREEN, BLACK", firstErr.Field())
		default:
			return fmt.Sprintf("%s is invalid", firstErr.Field())
		}
	}
	return "validation failed"
}

func sendErrorResponse(c *fiber.Ctx, status int, message string, traceID string) error {
	return c.Status(status).JSON(models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: message,
			TraceID: traceID,
		},
	})
}
