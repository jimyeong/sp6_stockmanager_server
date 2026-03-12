package controller

import (
	"net/http"

	"time"

	"github.com/jimyeongjung/owlverload_api/firebase"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
	"github.com/jimyeongjung/owlverload_api/v1/response"
	"github.com/jimyeongjung/owlverload_api/v1/service"
)

type GetProductWithExpiringStocksBetweenDaysRequest struct {
	StartPeriod time.Time `json:"start_period"`
	EndPeriod   time.Time `json:"end_period"`
}
type GetProductWithExpiringStocksBetweenDaysResponse struct {
	Products  []v1.Product `json:"products"`
	StartDate string       `json:"start_date"`
	EndDate   string       `json:"end_date"`
	Total     int          `json:"total"`
	Message   string       `json:"message"`
}

// GET /products/expiring-stocks?startDate={startDate}&endDate={endDate}
func HandleGetProductsWithExpiringStocksByDateRange(w http.ResponseWriter, r *http.Request) {
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	if userEmail == "" {
		response.WriteV1ServiceError(w, "User authentication required", false, http.StatusUnauthorized)
		return
	}
	queryParams := r.URL.Query()
	startDate := queryParams.Get("startDate") // yyyy-mm-dd
	endDate := queryParams.Get("endDate")     // yyyy-mm-dd

	if startDate == "" || endDate == "" {
		response.WriteV1ServiceError(w, "Invalid startDate or endDate parameter", false, http.StatusBadRequest)
		return
	}

	startDateTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		response.WriteV1ServiceError(w, "Invalid startDate parameter", false, http.StatusBadRequest)
		return
	}
	endDateTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		response.WriteV1ServiceError(w, "Invalid endDate parameter", false, http.StatusBadRequest)
		return
	}

	// validate startDate and endDate
	if startDateTime.After(endDateTime) {
		response.WriteV1ServiceError(w, "Start date cannot be greater than end date", false, http.StatusBadRequest)
		return
	}
	startDateString := startDateTime.Format("2006-01-02")
	endDateString := endDateTime.Format("2006-01-02")

	products, err := service.GetProductsWithExpiringStockWithinRange(startDateString, endDateString)
	if err != nil {
		response.WriteV1ServiceError(w, "Failed to get products with expiring stocks between days", false, http.StatusInternalServerError)
		return
	}
	result := GetProductWithExpiringStocksBetweenDaysResponse{
		Products:  products,
		StartDate: startDateString,
		EndDate:   endDateString,
		Total:     len(products),
		Message:   "Products with expiring stocks between days retrieved successfully",
	}

	response.WriteV1ServiceResponse(w, response.V1ServiceResponse[GetProductWithExpiringStocksBetweenDaysResponse]{
		Message: "Products with expiring stocks between days retrieved successfully",
		Payload: result,
		Success: true,
	})
}
