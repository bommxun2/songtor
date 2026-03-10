package main

import "time"

type DivertRequest struct {
	NotificationID    string `json:"notification_id" validate:"required"`
	RequestID         string `json:"request_id" validate:"required"`
	CurrentHospitalID string `json:"current_hospital_id" validate:"required"`
	NewHospitalID     string `json:"new_hospital_id" validate:"required,nefield=CurrentHospitalID"`
	Reason            string `json:"reason" validate:"required"`
	RequestedBy       string `json:"requested_by" validate:"required"`
}

type DivertResponse struct {
	NotificationID string `json:"notification_id"`
	RequestID      string `json:"request_id"`
	Status         string `json:"status"`
	NewHospitalID  string `json:"new_hospital_id,omitempty"`
	ReasonCode     string `json:"reason_code,omitempty"`
	ReasonMessage  string `json:"reason_message,omitempty"`
}

type ProcessedMessage struct {
	MessageID string `gorm:"primaryKey;type:varchar(50)"`
	CreatedAt int64  `gorm:"autoCreateTime"`
}

type NotificationRecord struct {
	ID             string      `gorm:"primaryKey;column:notification_id;type:varchar(50)"`
	HospitalID     string      `gorm:"column:hospital_id;type:varchar(50);not null"`
	AmbulanceID    string      `gorm:"column:ambulance_id;type:varchar(50);not null"`
	ArrivalTime    *time.Time  `gorm:"column:arrival_time;type:datetime"`
	Status         string      `gorm:"column:status;type:enum('EN_ROUTE','ARRIVED','HANDOVER_COMPLETED','CANCELLED');not null;default:'EN_ROUTE'"`
	IdempotencyKey string      `gorm:"column:idempotency_key;type:char(36);not null;uniqueIndex"`
	CreatedAt      time.Time   `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time   `gorm:"column:updated_at;autoUpdateTime"`
	PatientData    PatientData `gorm:"foreignKey:NotificationID;references:ID"`
}

type PatientData struct {
	NotificationID string  `gorm:"primaryKey;column:notification_id;type:varchar(50)"`
	TriageLevel    string  `gorm:"column:triage_level;type:enum('RED','YELLOW','GREEN','BLACK');not null"`
	Symptom        string  `gorm:"column:symptom;type:text;not null"`
	PatientAge     *int    `gorm:"column:patient_age"`
	PatientGender  *string `gorm:"column:patient_gender;type:enum('M','F','U')"`
	Bp             *string `gorm:"column:bp;type:varchar(20)"`
	Hr             *int    `gorm:"column:hr"`
	Spo2           *int    `gorm:"column:spo2"`
	AttachmentURLs string  `gorm:"column:attachment_urls;type:json"`
}
