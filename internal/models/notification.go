package models

type PatientInfo struct {
	Age    int    `json:"age"`
	Gender string `json:"gender"`
}

type Vitals struct {
	Bp   string `json:"bp"`
	Hr   int    `json:"hr"`
	Spo2 int    `json:"spo2"`
}

type CreateNotificationRequest struct {
	HospitalID     string      `json:"hospital_Id" validate:"required"`
	AmbulanceID    string      `json:"ambulance_Id"`
	PatientInfo    PatientInfo `json:"patient_info"`
	TriageLevel    string      `json:"triage_level" validate:"required,oneof=RED YELLOW GREEN BLACK"`
	Symptom        string      `json:"symptom" validate:"required"`
	Vitals         Vitals      `json:"vitals"`
	AttachmentURLs []string    `json:"attachment_urls"`
}
