package models

import (
	"encoding/json"
	"net/http"
)

type ServiceResponse struct {
	Message    string      `json:"message"`
	Payload    interface{} `json:"payload"`
	Success    bool        `json:"success"`
	UserExists bool        `json:"userExists"`
}

func WriteServiceResponse(w http.ResponseWriter, message string, data interface{}, success bool, userExists bool, responseCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	return json.NewEncoder(w).Encode(ServiceResponse{
		Message:    message,
		Payload:    data,
		Success:    success,
		UserExists: userExists,
	})
}

func WriteServiceError(w http.ResponseWriter, message string, success bool, userExists bool, errorCode int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(errorCode)
	return json.NewEncoder(w).Encode(ServiceResponse{
		Message:    message,
		Payload:    nil,
		Success:    success,
		UserExists: userExists,
	})
}
