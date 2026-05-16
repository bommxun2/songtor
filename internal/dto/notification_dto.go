package dto

// Requset body for creating a notification
type EmergencyRequest struct {
	HospitalID     string      `json:"hospital_Id"`
	AmbulanceID    string      `json:"ambulance_Id"`
	PatientInfo    PatientInfo `json:"patient_info"`
	TriageLevel    string      `json:"triage_level"`
	Symptom        string      `json:"symptom"`
	Vitals         Vitals      `json:"vitals"`
	AttachmentURLs []string    `json:"attachment_urls"`
}

type PatientInfo struct {
	FirstName      string `json:"first_name"`
	LastName       string `json:"last_name"`
	Gender         string `json:"gender"`
	Job            string `json:"job"`
	FoundLocation  string `json:"found_location"`
	AgeCategory    string `json:"age_category"`
	PhysicalDesc   string `json:"physical_desc"`
	PhysicalRemark string `json:"physical_remark"`
	ClothesDesc    string `json:"clothes_desc"`
	LifeStatus     string `json:"life_status"`
}

type Vitals struct {
	BP   string `json:"bp"`
	HR   int    `json:"hr"`
	SpO2 int    `json:"spo2"`
}

type ListNotificationResponse struct {
	HospitalID string                 `json:"hospital_id"`
	Items      []NotificationListItem `json:"items"`
}

type NotificationListItem struct {
	NotificationID string          `json:"notification_id"`
	AmbulanceID    string          `json:"ambulance_id"`
	PatientInfo    ListPatientInfo `json:"patient_info"`
	TriageLevel    string          `json:"triage_level"`
	Symptom        string          `json:"symptom"`
	Vitals         Vitals          `json:"vitals"`
	AttachmentURLs []string        `json:"attachment_urls"`
	Status         string          `json:"status"`
}

type ListPatientInfo struct {
	Age    int    `json:"age,omitempty"`
	Gender string `json:"gender,omitempty"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}
