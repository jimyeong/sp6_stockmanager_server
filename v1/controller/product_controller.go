package controller

import (
	"encoding/json"
	"net/http"

	"github.com/jimyeongjung/owlverload_api/firebase"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
	"github.com/jimyeongjung/owlverload_api/v1/response"
	"github.com/jimyeongjung/owlverload_api/v1/service"
)

func HandleProductUpdate(w http.ResponseWriter, r *http.Request) {
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	if tokenClaims.Email == "" {
		response.WriteV1ServiceError(w, "User authentication required", false, http.StatusUnauthorized)
		return
	}

	var product v1.Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		response.WriteV1ServiceError(w, "Failed to decode request body", false, http.StatusBadRequest)
		return
	}

	//check values are null or empty
	if product.ID == "" {
		response.WriteV1ServiceError(w, "Product ID is required", false, http.StatusBadRequest)
		return
	}
	if product.Code == "" {
		response.WriteV1ServiceError(w, "Product Code is required", false, http.StatusBadRequest)
		return
	}
	if product.Name == "" && product.NameJpn == "" && product.NameChn == "" && product.NameKor == "" && product.NameEng == "" {
		response.WriteV1ServiceError(w, "Product Name, Type, NameJpn, NameChn, NameKor, NameEng are required", false, http.StatusBadRequest)
		return
	}
	if product.ImagePath == "" {
		response.WriteV1ServiceError(w, "Product ImagePath is required", false, http.StatusBadRequest)
		return
	}
	if product.Price == 0 {
		response.WriteV1ServiceError(w, "Product Price is required", false, http.StatusBadRequest)
		return
	}

	err = service.UpdateProductService(product)
	if err != nil {
		response.WriteV1ServiceError(w, "Failed to update product", false, http.StatusInternalServerError)
		return
	}
	response.WriteV1ServiceResponse(w, response.V1ServiceResponse[v1.Product]{
		Message: "Product updated successfully",
		Payload: product,
		Success: true,
	})
}
