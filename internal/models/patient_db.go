package models

type PatientData struct {
	NotificationID string   `gorm:"primaryKey;column:notification_id"`
	TriageLevel    string   `gorm:"column:triage_level;not null"`
	Symptom        string   `gorm:"column:symptom;not null"`
	CategoryAge    *string  `gorm:"column:category_age"`
	Gender         *string  `gorm:"column:gender"`
	Bp             *string  `gorm:"column:bp"`
	Hr             *int     `gorm:"column:hr"`
	Spo2           *int     `gorm:"column:spo2"`
	AttachmentURLs []string `gorm:"column:attachment_urls;type:json;serializer:json"`
}
