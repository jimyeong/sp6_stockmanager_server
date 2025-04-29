package models

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
)

type ServiceResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type ServiceResponseWithData[T any] struct {
	ServiceResponse
	Data T `json:"data"`
}

type ServiceResponseWithLength[T any] struct {
	ServiceResponseWithData[T]
	dataLength int `json:"length"`
}

type ServiceResponseWithLengthAndUserState[T any] struct {
	ServiceResponseWithLength[T]
	UserExists bool `json:"userExists"`
}

func CreateServiceResponse[T any](message string, statusCode int, data T, userExists bool) ServiceResponseWithLengthAndUserState[T] {
	fmt.Println("---CreateServiceResponse---")
	fmt.Println("message:", message)
	serviceResponse := ServiceResponse{
		Code:    statusCode,
		Message: message,
	}
	serviceResponseWithData := ServiceResponseWithData[T]{
		ServiceResponse: serviceResponse,
		Data:            data,
	}
	var serviceResponseWithLength ServiceResponseWithLength[T]
	if reflect.TypeOf(data).Kind() == reflect.Array || reflect.TypeOf(data).Kind() == reflect.Slice {
		// get length of data
		serviceResponseWithLength = ServiceResponseWithLength[T]{
			ServiceResponseWithData: serviceResponseWithData,
			dataLength:              reflect.ValueOf(data).Len(),
		}
	} else {
		serviceResponseWithLength = ServiceResponseWithLength[T]{
			ServiceResponseWithData: serviceResponseWithData,
			dataLength:              1,
		}
	}
	serviceResponseWithLengthAndUserState := ServiceResponseWithLengthAndUserState[T]{
		ServiceResponseWithLength: serviceResponseWithLength,
		UserExists:                userExists,
	}
	return serviceResponseWithLengthAndUserState
}

func WriteServiceResponse[T any](w http.ResponseWriter, response ServiceResponseWithLengthAndUserState[T], userExists bool) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	return json.NewEncoder(w).Encode(response)
}

func WriteServiceError[T any](w http.ResponseWriter, response ServiceResponseWithLengthAndUserState[T], userExists bool) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	return json.NewEncoder(w).Encode(response)
}
