package models

import (
	"time"
)

type OutboxEvent struct {
	ID        uint   `gorm:"primaryKey"`
	TopicARN  string `gorm:"not null"`
	Payload   string `gorm:"type:text;not null"`
	Status    string `gorm:"type:varchar(20);default:'PENDING'"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
