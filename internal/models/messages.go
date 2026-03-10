package models

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
