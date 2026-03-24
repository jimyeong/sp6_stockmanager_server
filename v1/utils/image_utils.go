package utils

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/disintegration/imaging"
	"golang.org/x/exp/slices"
)

// ImageProcessingConfig configures image processing (resize, quality, format).
type ImageProcessingConfig struct {
	MaxWidth  int
	Quality   int
	Format    string
	StripExif bool
}

// ProcessImage processes the image according to specifications
func ProcessImage(imageData []byte, config ImageProcessingConfig) ([]byte, error) {
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
func UploadToR2(imageData []byte, filename string) (string, error) {
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
func IsValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
	}

	return slices.Contains(validTypes, strings.ToLower(contentType))
}

// extractFilenameFromPath extracts the filename from a full URL or path
func ExtractFilenameFromPath(imagePath string) (string, error) {
	// If it's a full URL, parse it
	if strings.HasPrefix(imagePath, "http://") || strings.HasPrefix(imagePath, "https://") {
		parsedURL, err := url.Parse(imagePath)
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
func DeleteFromR2(filename string) error {
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
