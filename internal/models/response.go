package models

type SuccessResponse struct {
	NotificationID string `json:"notification_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	TraceID string `json:"trace_id"`
}

type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}
