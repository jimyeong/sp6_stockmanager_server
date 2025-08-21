package apis

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"github.com/google/uuid"
	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
	"golang.org/x/exp/slices"
)

type ImageUploadResponse struct {
	ImagePath string `json:"image_path"`
	ImageID   string `json:"image_id"`
	FileSize  int64  `json:"file_size"`
	Message   string `json:"message"`
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
	FileName  string `json:"file_name"`
}

type ImageDeleteRequest struct {
	ImagePath string `json:"image_path"`
}

type ImageDeleteResponse struct {
	Message   string `json:"message"`
	Success   bool   `json:"success"`
	Timestamp string `json:"timestamp"`
	ImagePath string `json:"image_path"`
}

type ImageProcessingConfig struct {
	MaxWidth  int
	Quality   int
	Format    string
	StripExif bool
}

// Initialize image processing (no initialization needed for imaging library)

// HandleImageUpload handles POST requests to upload and process images
func HandleImageUpload(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Printf("User uploading image: %s\n", userEmail)

	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(32 << 20) // 32MB max memory
	if err != nil {
		log.Printf("Error parsing multipart form: %v", err)
		models.WriteServiceError(w, "Failed to parse form data", false, true, http.StatusBadRequest)
		return
	}

	// Get the uploaded file
	file, header, err := r.FormFile("image")
	if err != nil {
		log.Printf("Error retrieving file from form: %v", err)
		models.WriteServiceError(w, "No image file provided", false, true, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
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

	// Process the image
	processedData, err := processImage(fileData, ImageProcessingConfig{
		MaxWidth:  600,
		Quality:   70,
		Format:    "jpeg",
		StripExif: true,
	})
	if err != nil {
		log.Printf("Error processing image: %v", err)
		models.WriteServiceError(w, "Failed to process image", false, true, http.StatusInternalServerError)
		return
	}

	// Generate UUID for the image
	imageID := uuid.New().String()

	// Create filename with UUID
	filename := fmt.Sprintf("%s.jpg", imageID)
	fmt.Println("filename@@@@@@@@@@@@@@@@@@@@@", filename)

	// Upload to R2 Cloudflare
	imagePath, err := uploadToR2(processedData, filename)
	if err != nil {
		log.Printf("Error uploading to R2: %v", err)
		models.WriteServiceError(w, "Failed to upload image to storage", false, true, http.StatusInternalServerError)
		return
	}
	// fmt.Println("imageURL@@@@@@@@@@@@@@@@@@@@@", imageURL)

	// Prepare response
	response := ImageUploadResponse{
		ImagePath: imagePath,
		ImageID:   imageID,
		FileSize:  int64(len(processedData)),
		Message:   "Image uploaded successfully",
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
		FileName:  filename,
	}

	models.WriteServiceResponse(w, "Image uploaded successfully", response, true, true, http.StatusOK)
}

// processImage processes the image according to specifications
func processImage(imageData []byte, config ImageProcessingConfig) ([]byte, error) {
	// Decode image from bytes
	reader := bytes.NewReader(imageData)
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// Get original dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Resize if necessary, maintaining aspect ratio
	var resizedImg image.Image = img
	if width > config.MaxWidth {
		newHeight := int(float64(height) * float64(config.MaxWidth) / float64(width))
		resizedImg = imaging.Resize(img, config.MaxWidth, newHeight, imaging.Lanczos)
	}

	// Convert to JPEG and encode with specified quality
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: config.Quality}

	err = jpeg.Encode(&buf, resizedImg, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to encode processed image: %v", err)
	}

	log.Printf("Image processed: %s -> JPEG, original: %dx%d, final: %dx%d",
		format, width, height, resizedImg.Bounds().Dx(), resizedImg.Bounds().Dy())

	return buf.Bytes(), nil
}

// uploadToR2 uploads the processed image to Cloudflare R2
func uploadToR2(imageData []byte, filename string) (string, error) {
	// Get R2 configuration from environment variables
	r2AccessKey := os.Getenv("R2_ACCESS_KEY_ID")
	r2SecretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	r2Endpoint := os.Getenv("R2_ENDPOINT")
	r2BucketName := os.Getenv("R2_BUCKET_NAME")
	r2PublicDomain := os.Getenv("R2_PUBLIC_DOMAIN")

	if r2AccessKey == "" || r2SecretKey == "" || r2Endpoint == "" || r2BucketName == "" {
		return "", fmt.Errorf("R2 configuration missing. Please set R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_ENDPOINT, and R2_BUCKET_NAME environment variables")
	}

	// Create custom AWS config for R2
	r2Config := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(r2AccessKey, r2SecretKey, ""),
		Region:      "auto",
	}

	// Create S3 client for R2 with custom endpoint
	s3Client := s3.NewFromConfig(r2Config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
		o.UsePathStyle = true
	})

	// Create the put object input
	putObjectInput := &s3.PutObjectInput{
		Bucket:        aws.String(r2BucketName),
		Key:           aws.String(filename),
		Body:          bytes.NewReader(imageData),
		ContentType:   aws.String("image/jpeg"),
		ContentLength: aws.Int64(int64(len(imageData))),
	}

	// Upload the file
	_, err := s3Client.PutObject(context.TODO(), putObjectInput)
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %v", err)
	}

	// Construct public URL
	var imageURL string
	if r2PublicDomain != "" {
		imageURL = fmt.Sprintf("https://%s/%s", r2PublicDomain, filename)
	} else {
		imageURL = fmt.Sprintf("%s/%s/%s", r2Endpoint, r2BucketName, filename)
	}

	return imageURL, nil
}

// isValidImageType checks if the content type is a supported image format
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}

	return slices.Contains(validTypes, strings.ToLower(contentType))
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
	filename, err := extractFilenameFromPath(imagePath)
	if err != nil {
		log.Printf("Error extracting filename from path %s: %v", imagePath, err)
		models.WriteServiceError(w, "Invalid image path format", false, true, http.StatusBadRequest)
		return
	}

	// Delete from R2 Cloudflare
	err = deleteFromR2(filename)
	if err != nil {
		log.Printf("Error deleting from R2: %v", err)
		models.WriteServiceError(w, "Failed to delete image from storage", false, true, http.StatusInternalServerError)
		return
	}

	// Prepare response
	response := ImageDeleteResponse{
		Message:   "Image deleted successfully",
		Success:   true,
		Timestamp: time.Now().Format(time.RFC3339),
		ImagePath: imagePath,
	}

	models.WriteServiceResponse(w, "Image deleted successfully", response, true, true, http.StatusOK)
}

// extractFilenameFromPath extracts the filename from a full URL or path
func extractFilenameFromPath(imagePath string) (string, error) {
	// If it's a full URL, parse it
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		parsedURL, err := url.Parse(imagePath)
		fmt.Println("parsedURL@@@@@@@@@@@@@@@@@@@@@", parsedURL)
		fmt.Println("imagePath@@@@@@@@@@@@@@@@@@@@@", imagePath)
		if err != nil {
			return "", fmt.Errorf("invalid URL format: %v", err)
		}
		// Extract filename from URL path
		filename := path.Base(parsedURL.Path)
		if filename == "." || filename == "/" {
			return "", fmt.Errorf("no filename found in URL path")
		}
		return filename, nil
	}

	// If it's just a path, extract the base filename
	filename := path.Base(imagePath)
	if filename == "." || filename == "/" {
		return "", fmt.Errorf("no filename found in path")
	}

	return filename, nil
}

// deleteFromR2 deletes an image from Cloudflare R2 storage
func deleteFromR2(filename string) error {
	// Get R2 configuration from environment variables
	r2AccessKey := os.Getenv("R2_ACCESS_KEY_ID")
	r2SecretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	r2Endpoint := os.Getenv("R2_ENDPOINT")
	r2BucketName := os.Getenv("R2_BUCKET_NAME")

	if r2AccessKey == "" || r2SecretKey == "" || r2Endpoint == "" || r2BucketName == "" {
		return fmt.Errorf("R2 configuration missing. Please set R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_ENDPOINT, and R2_BUCKET_NAME environment variables")
	}

	// Create custom AWS config for R2
	r2Config := aws.Config{
		Credentials: credentials.NewStaticCredentialsProvider(r2AccessKey, r2SecretKey, ""),
		Region:      "auto",
	}

	// Create S3 client for R2 with custom endpoint
	s3Client := s3.NewFromConfig(r2Config, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(r2Endpoint)
		o.UsePathStyle = true
	})

	// Create the delete object input
	deleteObjectInput := &s3.DeleteObjectInput{
		Bucket: aws.String(r2BucketName),
		Key:    aws.String("images/" + filename),
	}

	// Delete the file
	_, err := s3Client.DeleteObject(context.TODO(), deleteObjectInput)
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %v", err)
	}

	log.Printf("Successfully deleted image: %s", filename)
	return nil
}
