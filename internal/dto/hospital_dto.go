package dto

// Response from hospital resource service
type HospitalResourceResponse struct {
	HospitalID string     `json:"hospitalId"`
	Resources  []Resource `json:"resources"`
}

type Resource struct {
	ResourceType      string `json:"resourceType"`
	TotalCapacity     int    `json:"totalCapacity"`
	AvailableQuantity int    `json:"availableQuantity"`
	ResourceStatus    string `json:"resourceStatus"`
	LastUpdatedTime   string `json:"lastUpdatedTime"`
}
