package apis

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
	"github.com/jimyeongjung/owlverload_api/utils"
)

// BarcodeRequest defines the structure for the barcode save request
type BarcodeRequest struct {
	Barcode string `json:"barcode"`
}

// HandleSaveBarcode handles POST requests to save a barcode to the database
func HandleSaveBarcode(w http.ResponseWriter, r *http.Request) {
	defer utils.Trace()()
	utils.Info("Starting HandleSaveBarcode operation")

	// Get authenticated user from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	utils.Info("User handling barcode save: %s", userEmail)

	if userEmail == "" {
		utils.Warn("Unauthorized barcode save attempt - missing user email")
		models.WriteServiceError(w, "User authentication required", false, false, http.StatusUnauthorized)
		return
	}

	// Read and parse request body
	utils.Debug("Reading request body")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Error("Failed to read request body: %v", err)
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var request BarcodeRequest
	utils.Debug("Unmarshaling request JSON")
	err = json.Unmarshal(body, &request)
	if err != nil {
		utils.Error("Invalid request format: %v", err)
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	utils.Debug("Validating request")
	if request.Barcode == "" {
		utils.Warn("Missing barcode in request")
		models.WriteServiceError(w, "Barcode is required", false, true, http.StatusBadRequest)
		return
	}

	// Save the barcode
	utils.Info("Saving barcode: %s", request.Barcode)
	// Still passing userEmail for logging purposes even though we don't store it anymore
	barcode, err := models.SaveBarcode(request.Barcode, userEmail)
	if err != nil {
		utils.Error("Error saving barcode: %v", err)
		models.WriteServiceError(w, err.Error(), false, true, http.StatusInternalServerError)
		return
	}

	// Prepare response
	utils.Info("Barcode saved successfully: %s", barcode.Code)
	payload := map[string]interface{}{
		"barcode": barcode.Code,
		"message": "barcode saved",
	}

	// Return success response
	models.WriteServiceResponse(w, "Barcode saved successfully", payload, true, true, http.StatusOK)
}