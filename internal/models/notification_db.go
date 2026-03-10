package models

import "time"

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
