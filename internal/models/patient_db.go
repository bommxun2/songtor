package models

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
