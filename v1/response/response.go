package response

import (
	"encoding/json"
	"net/http"
)

type V1ServiceResponse[T any] struct {
	Message string `json:"message"`
	Payload T      `json:"payload"`
	Success bool   `json:"success"`
}
type V1PagenatedServiceResponse[T any] struct {
	V1ServiceResponse[[]T]
	Total      int  `json:"total"`
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	HasNext    bool `json:"hasNext"`
	HasPrev    bool `json:"hasPrev"`
	TotalPages int  `json:"totalPages"`
}

func WriteV1ServiceResponse[T any](w http.ResponseWriter, response V1ServiceResponse[T]) error {
	return json.NewEncoder(w).Encode(response)
}

func WriteV1PagenatedServiceResponse[T any](w http.ResponseWriter, response V1PagenatedServiceResponse[T]) error {
	return json.NewEncoder(w).Encode(response)
}

func WriteV1ServiceError(w http.ResponseWriter, message string, success bool, errorCode int) error {
	return json.NewEncoder(w).Encode(V1ServiceResponse[interface{}]{
		Message: message,
		Payload: nil,
		Success: success,
	})
}
