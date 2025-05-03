package models

// package models

// import (
// 	"encoding/json"
// 	"net/http"
// 	"reflect"
// )

// type ServiceResponse struct {
// 	Message string `json:"message"`
// }

// type ServiceResponseWithData[T any] struct {
// 	ServiceResponse
// 	Data T `json:"payload"`
// }

// type ServiceResponseWithLength[T any] struct {
// 	ServiceResponseWithData[T]
// 	dataLength int `json:"length"`
// }

// type ServiceResponseWithLengthAndUserState[T any] struct {
// 	ServiceResponseWithLength[T]
// 	UserExists bool `json:"userExists"`
// }

// func CreateServiceResponse[T any](message string, data T, userExists bool) ServiceResponseWithLengthAndUserState[T] {
// 	serviceResponse := ServiceResponse{
// 		Message: message,
// 	}
// 	serviceResponseWithData := ServiceResponseWithData[T]{
// 		ServiceResponse: serviceResponse,
// 		Data:            data,
// 	}
// 	var serviceResponseWithLength ServiceResponseWithLength[T]
// 	if reflect.TypeOf(data).Kind() == reflect.Array || reflect.TypeOf(data).Kind() == reflect.Slice {
// 		// get length of data
// 		serviceResponseWithLength = ServiceResponseWithLength[T]{
// 			ServiceResponseWithData: serviceResponseWithData,
// 			dataLength:              reflect.ValueOf(data).Len(),
// 		}
// 	} else {
// 		serviceResponseWithLength = ServiceResponseWithLength[T]{
// 			ServiceResponseWithData: serviceResponseWithData,
// 			dataLength:              1,
// 		}
// 	}
// 	serviceResponseWithLengthAndUserState := ServiceResponseWithLengthAndUserState[T]{
// 		ServiceResponseWithLength: serviceResponseWithLength,
// 		UserExists:                userExists,
// 	}
// 	return serviceResponseWithLengthAndUserState
// }

// func WriteServiceResponse[T any](w http.ResponseWriter, message string, data T, userExists bool) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(
// 		CreateServiceResponse(message, data, userExists),
// 	)
// }

// func WriteServiceError(w http.ResponseWriter, message string, userExists bool) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusInternalServerError)
// 	payload := make([]string, 0)
// 	json.NewEncoder(w).Encode(
// 		CreateServiceResponse(message, payload, userExists),
// 	)
// }
