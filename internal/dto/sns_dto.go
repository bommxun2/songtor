package dto

// Requset for sending sns to tracemissing service
type PatientTransferRequest struct {
	DestinationHospitalID string          `json:"destination_hospital_id"`
	Characteristics       Characteristics `json:"characteristics"`
	Media                 Media           `json:"media"`
	ReportedBy            string          `json:"reported_by"`
}

type Characteristics struct {
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

type Media struct {
	MissingPersonPhoto string `json:"missing_person_photo"`
}
