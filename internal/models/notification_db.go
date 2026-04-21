package models

import "time"

type NotificationRecord struct {
	ID             string      `gorm:"primaryKey;column:notification_id"`
	HospitalID     string      `gorm:"column:hospital_id;not null"`
	AmbulanceID    string      `gorm:"column:ambulance_id;not null"`
	ArrivalTime    *time.Time  `gorm:"column:arrival_time"`
	Status         string      `gorm:"column:status;not null;default:'EN_ROUTE'"`
	IdempotencyKey string      `gorm:"column:idempotency_key;not null;uniqueIndex"`
	CreatedAt      time.Time   `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt      time.Time   `gorm:"column:updated_at;autoUpdateTime"`
	PatientData    PatientData `gorm:"foreignKey:NotificationID;references:ID"`
}
