package service

import (
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	controllerModels "github.com/jimyeongjung/owlverload_api/v1/controller/models"
	"github.com/jimyeongjung/owlverload_api/v1/utils"
)

func UploadImageService(fileData []byte) (controllerModels.ImageUploadResponse, error) {

	// service
	// Process the image
	processedData, err := utils.ProcessImage(fileData, utils.ImageProcessingConfig{
		MaxWidth:  600,
		Quality:   70,
		Format:    "jpeg",
		StripExif: true,
	})
	if err != nil {
		log.Printf("Error processing image: %v", err)
		return controllerModels.ImageUploadResponse{}, err
	}
	// Generate UUID for the image
	imageID := uuid.New().String()

	// Create filename with UUID
	filename := fmt.Sprintf("%s.jpg", imageID)
	// Upload to R2 Cloudflare
	imagePath, err := utils.UploadToR2(processedData, filename)
	if err != nil {
		log.Printf("Error uploading to R2: %v", err)

		return controllerModels.ImageUploadResponse{}, err
	}

	// Prepare response
	filename = "/" + filename
	result := controllerModels.ImageUploadResponse{
		ImagePath: imagePath,
		ImageID:   imageID,
		FileSize:  int64(len(processedData)),
		Message:   "Image uploaded successfully",
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
		FileName:  filename,
	}

	return result, nil
}
