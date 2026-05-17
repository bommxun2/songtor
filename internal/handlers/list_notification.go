package handlers

import (
	"fmt"
	"songtor/internal/dto"
	"songtor/internal/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ListNotificationHandler struct {
	db *gorm.DB
}

func NewListNotificationHandler(db *gorm.DB) *ListNotificationHandler {
	return &ListNotificationHandler{db: db}
}

func (h *ListNotificationHandler) ListNotifications(c *fiber.Ctx) error {
	hospitalID := c.Params("hospital_id")
	if hospitalID == "" {
		fmt.Println("Validation error: hospital_id is required")
		return c.Status(fiber.StatusBadRequest).JSON(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "VALIDATION_ERROR",
				Message: "hospital_id required",
				TraceID: uuid.New().String(),
			},
		})
	}

	status := c.Query("status", "EN_ROUTE")

	var notifications []models.NotificationRecord
	err := h.db.Preload("PatientData").
		Where("hospital_id = ? AND status = ?", hospitalID, status).
		Find(&notifications).Error

	if err != nil {
		fmt.Printf("Failed to fetch notifications: %v\n", err)
		return c.Status(fiber.StatusInternalServerError).JSON(dto.ErrorResponse{
			Error: dto.ErrorDetail{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to fetch notifications",
				TraceID: uuid.New().String(),
			},
		})
	}

	items := make([]dto.NotificationListItem, 0)
	for _, n := range notifications {

		ageInt := 0
		if n.PatientData.CategoryAge != nil {
			// Try to parse age from category_age if possible, else 0.
			// The API contract shows age as 45, which means maybe some systems store age inside CategoryAge or we just don't have exact age.
			parsed, err := strconv.Atoi(*n.PatientData.CategoryAge)
			if err == nil {
				ageInt = parsed
			}
		}

		genderStr := ""
		if n.PatientData.Gender != nil {
			genderStr = *n.PatientData.Gender
		}

		bpStr := ""
		if n.PatientData.Bp != nil {
			bpStr = *n.PatientData.Bp
		}

		hrInt := 0
		if n.PatientData.Hr != nil {
			hrInt = *n.PatientData.Hr
		}

		spo2Int := 0
		if n.PatientData.Spo2 != nil {
			spo2Int = *n.PatientData.Spo2
		}

		items = append(items, dto.NotificationListItem{
			NotificationID: n.ID,
			AmbulanceID:    n.AmbulanceID,
			PatientInfo: dto.ListPatientInfo{
				Age:    ageInt,
				Gender: genderStr,
			},
			TriageLevel: n.PatientData.TriageLevel,
			Symptom:     n.PatientData.Symptom,
			Vitals: dto.Vitals{
				BP:   bpStr,
				HR:   hrInt,
				SpO2: spo2Int,
			},
			AttachmentURLs: n.PatientData.AttachmentURLs,
			Status:         n.Status,
		})
	}

	response := dto.ListNotificationResponse{
		HospitalID: hospitalID,
		Items:      items,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
