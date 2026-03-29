package controller

import (
	"fmt"
	"net/http"

	"encoding/json"

	"github.com/jimyeongjung/owlverload_api/firebase"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
	v1Models "github.com/jimyeongjung/owlverload_api/v1/models"
	"github.com/jimyeongjung/owlverload_api/v1/response"
	service "github.com/jimyeongjung/owlverload_api/v1/service"
)

func HandleCreateStock(w http.ResponseWriter, r *http.Request) {
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	if userEmail == "" {
		response.WriteV1ServiceError(w, "User authentication required", false, http.StatusUnauthorized)
		return
	}

	var req CreateStockRequest
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteV1ServiceError(w, "Invalid request body", false, http.StatusBadRequest)
		return
	}
	if req.ItemID == "" && req.Code == "" {
		response.WriteV1ServiceError(w, "Item ID or Code is required", false, http.StatusBadRequest)
		return
	}
	if req.Quantity <= 0 {
		response.WriteV1ServiceError(w, "Quantity must be greater than 0", false, http.StatusBadRequest)
		return
	}
	if req.StockType == "" {
		response.WriteV1ServiceError(w, "Stock type is required", false, http.StatusBadRequest)
		return
	}

	stock := v1Models.Stock{
		ItemId:            req.ItemID,
		StockType:         req.StockType,
		ExpiryDate:        req.ExpiryDate,
		Location:          req.Location,
		RegisteringPerson: userEmail,
		Notes:             req.Notes,
		DiscountRate:      req.DiscountRate,
		Quantity:          req.Quantity,
	}

	err := service.CreateStockService(r.Context(), stock, idempotencyKey)
	fmt.Println("@@@@@@@@@@@@@@@ERR1", err)
	if err != nil {

		response.WriteV1ServiceError(w, "Failed to create stock", false, http.StatusInternalServerError)
		return
	}

	createStockResponse := CreateStockResponse{
		Message:       "Stock created successfully",
		UpdatedStocks: []v1.Stock{stock},
		AddedStocks:   []v1.Stock{stock},
	}

	response.WriteV1ServiceResponse(w, response.V1ServiceResponse[CreateStockResponse]{
		Message: createStockResponse.Message,
		Payload: createStockResponse,
		Success: true,
	})
}
