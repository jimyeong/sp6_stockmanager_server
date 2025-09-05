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
	fmt.Println("---HANDLESAVEBARCODE---")

	// Get authenticated user from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("---USER EMAIL---", userEmail)

	if userEmail == "" {
		fmt.Println("---Unauthorized barcode save attempt - missing user email---")
		models.WriteServiceError(w, "User authentication required", false, false, http.StatusUnauthorized)
		return
	}

	// Read and parse request body
	fmt.Println("---Reading request body---")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("---Failed to read request body: %v---", err)
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var request BarcodeRequest
	fmt.Println("---Unmarshaling request JSON---")
	err = json.Unmarshal(body, &request)
	if err != nil {
		fmt.Println("---Invalid request format: %v---", err)
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	fmt.Println("---Validating request---")
	if request.Barcode == "" {
		fmt.Println("---Missing barcode in request---")
		models.WriteServiceError(w, "Barcode is required", false, true, http.StatusBadRequest)
		return
	}

	// Save the barcode
	// Still passing userEmail for logging purposes even though we don't store it anymore
	fmt.Println("---Saving barcode---")
	barcode, err := models.SaveBarcode(request.Barcode, userEmail)
	if err != nil {
		fmt.Println("---Error saving barcode: %v---", err)
		models.WriteServiceError(w, err.Error(), false, true, http.StatusInternalServerError)
		return
	}

	// Prepare response
	fmt.Println("---Preparing response---")
	payload := map[string]interface{}{
		"barcode": barcode.Code,
		"message": "barcode saved",
	}

	// Return success response
	fmt.Println("---Returning success response---")
	models.WriteServiceResponse(w, "Barcode saved successfully", payload, true, true, http.StatusOK)
}
