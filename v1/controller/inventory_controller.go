package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"encoding/json"

	"github.com/gorilla/mux"
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
type GetProductsWithStockWithDaysLeftRequest struct {
	DaysLeft int `json:"days_left"`
}
type GetProductsWithStockWithDaysLeftResponse struct {
	ExpiringProducts []v1.Product `json:"expiring_products"`
	DaysLeft         int          `json:"days_left"`
	Total            int          `json:"total"`
	Message          string       `json:"message"`
	IsExpired        bool         `json:"is_expired"`
}
type GetExpiredInventoryOlderThanDaysResponse struct {
	ProductsWithExpiredStocks []service.InventoryWithExpiredStock `json:"products_with_expired_stocks"`
	Total                     int                                 `json:"total"`
}
type GetInventoryByProductIdResponse struct {
	Products  map[string]*v1.Product `json:"products"`
	ProductId string                 `json:"product_id"`
	Total     int                    `json:"total"`
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

	products, err := service.GetInventoryOfExpiringStockWithinRange(startDateString, endDateString)
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
func HandleGetProductsWithStockWithDaysLeft(w http.ResponseWriter, r *http.Request) {
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	if userEmail == "" {
		response.WriteV1ServiceError(w, "User authentication required", false, http.StatusUnauthorized)
		return
	}
	queryParams := r.URL.Query()
	daysLeft := queryParams.Get("daysLeft")
	if daysLeft == "" {
		response.WriteV1ServiceError(w, "daysLeft parameter is required", false, http.StatusBadRequest)
		return
	}
	daysLeftInt, err := strconv.Atoi(daysLeft)
	if err != nil {
		response.WriteV1ServiceError(w, "Invalid daysLeft parameter", false, http.StatusBadRequest)
		return
	}
	products, err := service.GetInventoryWithDaysLeft(daysLeftInt)
	if err != nil {
		response.WriteV1ServiceError(w, "Failed to get products with stock with days left", false, http.StatusInternalServerError)
		return
	}
	result := GetProductsWithStockWithDaysLeftResponse{
		ExpiringProducts: products,
		DaysLeft:         daysLeftInt,
		Total:            len(products),
		IsExpired:        daysLeftInt < 0,
	}
	response.WriteV1ServiceResponse(w, response.V1ServiceResponse[GetProductsWithStockWithDaysLeftResponse]{
		Message: "Products with stock with days left retrieved successfully",
		Payload: result,
		Success: true,
	})
}

func HandleGetExpiredInventoryOlderThanDays(w http.ResponseWriter, r *http.Request) {

	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	if userEmail == "" {
		response.WriteV1ServiceError(w, "User authentication required", false, http.StatusUnauthorized)
		return
	}
	// olderThanDays
	queryParams := r.URL.Query()
	olderThanDays := queryParams.Get("olderThanDays")
	if olderThanDays == "" {
		response.WriteV1ServiceError(w, "olderThanDays parameter is required", false, http.StatusBadRequest)
		return
	}
	olderThanDaysInt, err := strconv.Atoi(olderThanDays)
	if err != nil {
		response.WriteV1ServiceError(w, "Invalid olderThanDays parameter", false, http.StatusBadRequest)
		return
	}
	if olderThanDaysInt < 0 {
		response.WriteV1ServiceError(w, "olderThanDays must be 0 or greater", false, http.StatusBadRequest)
		return
	}
	productsWithExpiredStocks, err := service.GetExpiredInventoryOlderThanDays(olderThanDaysInt)
	if err != nil {
		response.WriteV1ServiceError(w, "Failed to retrieve expired inventory", false, http.StatusInternalServerError)
		return
	}

	result := GetExpiredInventoryOlderThanDaysResponse{
		ProductsWithExpiredStocks: productsWithExpiredStocks,
		Total:                     len(productsWithExpiredStocks),
	}
	response.WriteV1ServiceResponse(w, response.V1ServiceResponse[GetExpiredInventoryOlderThanDaysResponse]{
		Message: "Expired inventory retrieved successfully",
		Payload: result,
		Success: true,
	})
}

func HandleGetInventoryByProductId(w http.ResponseWriter, r *http.Request) {
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	if tokenClaims.Email == "" {
		response.WriteV1ServiceError(w, "User authentication required", false, http.StatusUnauthorized)
		return
	}

	// apiRouter.HandleFunc("/products/inventory/{productId}", v1Controller.HandleGetInventoryByProductId).Methods("GET")
	productId := mux.Vars(r)["productId"]
	if productId == "" {
		response.WriteV1ServiceError(w, "productId is required", false, http.StatusBadRequest)
		return
	}

	productMap, err := service.GetInventoryByProductId(productId)
	if err != nil {
		response.WriteV1ServiceError(w, "Product not found", false, http.StatusNotFound)
		return
	}
	result := GetInventoryByProductIdResponse{
		Products:  productMap,
		ProductId: productId,
		Total:     len(productMap),
	}
	response.WriteV1ServiceResponse(w, response.V1ServiceResponse[GetInventoryByProductIdResponse]{
		Message: "Product inventory retrieved successfully",
		Payload: result,
		Success: true,
	})

}

func HandleFinaliseExpiredStock(w http.ResponseWriter, r *http.Request) {
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	if tokenClaims.Email == "" {
		response.WriteV1ServiceError(w, "User authentication required", false, http.StatusUnauthorized)
		return
	}

	var req FinaliseExpiredStockRequest
	// print request body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteV1ServiceError(w, "Invalid request body", false, http.StatusBadRequest)
		return
	}
	fmt.Println("-----HandleFinaliseExpiredStock-----", req)

	if req.StockId == "" {
		response.WriteV1ServiceError(w, "stockId is required", false, http.StatusBadRequest)
		return
	}
	req.PerformerEmail = tokenClaims.Email
	err := service.FinaliseExpiredStock(service.FinaliseExpiredStockParams{
		StockId:        req.StockId,
		EventType:      req.EventType,
		StockType:      req.StockType,
		PerformerEmail: req.PerformerEmail,
	})
	if err != nil {
		fmt.Println("-----Error-----", err)
		response.WriteV1ServiceError(w, "Failed to finalise expired stock", false, http.StatusInternalServerError)
		return
	}

	response.WriteV1ServiceResponse(w, response.V1ServiceResponse[string]{
		Message: "Expired stock finalised successfully",
		Success: true,
	})
}
