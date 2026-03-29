package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
)

// BarcodeRequest defines the structure for the barcode save request
type BarcodeRequest struct {
	Barcode string `json:"barcode"`
}

// HandleSaveBarcode handles POST requests to save a barcode to the database
func HandleSaveBarcode(w http.ResponseWriter, r *http.Request) {

	// Get authenticated user from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email

	if userEmail == "" {
		fmt.Println("---Unauthorized barcode save attempt - missing user email---")
		models.WriteServiceError(w, "User authentication required", false, false, http.StatusUnauthorized)
		return
	}

	// Read and parse request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var request BarcodeRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Barcode == "" {
		models.WriteServiceError(w, "Barcode is required", false, true, http.StatusBadRequest)
		return
	}

	// Save the barcode
	// Still passing userEmail for logging purposes even though we don't store it anymore
	barcode, err := models.SaveBarcode(request.Barcode, userEmail)
	if err != nil {
		models.WriteServiceError(w, err.Error(), false, true, http.StatusInternalServerError)
		return
	}

	// Prepare response
	payload := map[string]interface{}{
		"barcode": barcode.Code,
		"message": "barcode saved",
	}

	// Return success response
	models.WriteServiceResponse(w, "Barcode saved successfully", payload, true, true, http.StatusOK)
}
