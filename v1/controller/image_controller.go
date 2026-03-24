package controller

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
	controllerModels "github.com/jimyeongjung/owlverload_api/v1/controller/models"
	"github.com/jimyeongjung/owlverload_api/v1/response"
	"github.com/jimyeongjung/owlverload_api/v1/service"
	"github.com/jimyeongjung/owlverload_api/v1/utils"
)

func HandleImageUpload(w http.ResponseWriter, r *http.Request) {

	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Printf("User uploading image: %s\n", userEmail)

	if userEmail == "" {
		response.WriteV1ServiceError(w, "User authentication required", false, http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		response.WriteV1ServiceError(w, "Failed to parse form data", false, http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error retrieving file from form: %v", err)
		response.WriteV1ServiceError(w, "No image file provided", false, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !utils.IsValidImageType(contentType) {
		models.WriteServiceError(w, "Invalid image type. Only JPEG, PNG, and WebP are supported", false, true, http.StatusBadRequest)
		return
	}

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Printf("Error reading file data: %v", err)
		models.WriteServiceError(w, "Failed to read image file", false, true, http.StatusInternalServerError)
		return
	}

	// Upload image to R2 Cloudflare
	imageUploadResponse, err := service.UploadImageService(fileData)
	if err != nil {
		log.Printf("Error uploading image to R2: %v", err)
		models.WriteServiceError(w, "Failed to upload image to storage", false, true, http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := controllerModels.ImageUploadResponse{
		ImagePath: imageUploadResponse.ImagePath,
		ImageID:   imageUploadResponse.ImageID,
		FileSize:  imageUploadResponse.FileSize,
		Message:   imageUploadResponse.Message,
		Success:   imageUploadResponse.Success,
		Timestamp: imageUploadResponse.Timestamp,
		FileName:  imageUploadResponse.FileName,
	}

	models.WriteServiceResponse(w, "Image uploaded successfully", response, true, true, http.StatusOK)
}

// HandleImageUploadURL handles DELETE requests to remove images from R2 storage
func HandleImageDelete(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Printf("User deleting image: %s\n", userEmail)

	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Parse request body
	imagePath := r.URL.Query().Get("imagePath")
	fmt.Println("imagePath", imagePath)

	// Validate imagePath
	if imagePath == "" {
		models.WriteServiceError(w, "imagePath is required", false, true, http.StatusBadRequest)
		return
	}

	// Extract filename from the image path/URL
	filename, err := utils.ExtractFilenameFromPath(imagePath)
	if err != nil {
		log.Printf("Error extracting filename from path %s: %v", imagePath, err)
		models.WriteServiceError(w, "Invalid image path format", false, true, http.StatusBadRequest)
		return
	}

	// Delete from R2 Cloudflare
	err = utils.DeleteFromR2(filename)
	if err != nil {
		log.Printf("Error deleting from R2: %v", err)
		models.WriteServiceError(w, "Failed to delete image from storage", false, true, http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := controllerModels.ImageDeleteResponse{
		Message:   "Image deleted successfully",
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
		ImagePath: imagePath,
	}

	models.WriteServiceResponse(w, "Image deleted successfully", response, true, true, http.StatusOK)
}
