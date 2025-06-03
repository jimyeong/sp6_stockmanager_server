package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
	"github.com/jimyeongjung/owlverload_api/utils"
)

// VisionAnalyzeRequest defines the request structure for the client to server
type VisionAnalyzeRequest struct {
	Image string `json:"image"` // Base64 encoded image
}

// OpenAIVisionRequest defines the request structure for OpenAI API
type OpenAIVisionRequest struct {
	Model     string                `json:"model"`
	MaxTokens int                   `json:"max_tokens"`
	Messages  []OpenAIVisionMessage `json:"messages"`
}

type OpenAIVisionMessage struct {
	Role    string                       `json:"role"`
	Content []OpenAIVisionMessageContent `json:"content"`
}

type OpenAIVisionMessageContent struct {
	Type     string                       `json:"type"`
	Text     string                       `json:"text,omitempty"`
	ImageURL *OpenAIVisionMessageImageURL `json:"image_url,omitempty"`
}

type OpenAIVisionMessageImageURL struct {
	URL string `json:"url"`
}

// OpenAIResponse defines the response structure from OpenAI API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

// HandleVisionAnalyze handles the request to analyze a product image
func HandleVisionAnalyze(w http.ResponseWriter, r *http.Request) {
	defer utils.Trace()()
	utils.Info("Starting HandleVisionAnalyze operation")

	// Get authenticated user from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	utils.Info("User handling vision analysis: %s", userEmail)

	if userEmail == "" {
		utils.Warn("Unauthorized vision analysis attempt - missing user email")
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

	var request VisionAnalyzeRequest
	utils.Debug("Unmarshaling request JSON")
	err = json.Unmarshal(body, &request)
	if err != nil {
		utils.Error("Invalid request format: %v", err)
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	utils.Debug("Validating request")
	if request.Image == "" {
		utils.Warn("Missing image data in request")
		models.WriteServiceError(w, "Image data is required", false, true, http.StatusBadRequest)
		return
	}

	// Ensure image data is properly formatted
	imageData := request.Image
	if !strings.HasPrefix(imageData, "data:image/") && !strings.HasPrefix(imageData, "data:application/octet-stream;base64,") {
		// Assuming the image is just base64 encoded without the data URI scheme
		imageData = "data:image/jpeg;base64," + imageData
	}

	// Prepare OpenAI request
	utils.Debug("Preparing OpenAI request")
	openAIRequest := OpenAIVisionRequest{
		Model:     "gpt-4-vision-preview",
		MaxTokens: 600,
		Messages: []OpenAIVisionMessage{
			{
				Role: "user",
				Content: []OpenAIVisionMessageContent{
					{
						Type: "text",
						Text: "Analyze this product image and provide the following information in this exact format:\n1. Product Name:\n2. Expiry Date:\n3. Ingredients:\n4. Alcohol:\n5. Halal:\n6. Reasoning:",
					},
					{
						Type: "image_url",
						ImageURL: &OpenAIVisionMessageImageURL{
							URL: imageData,
						},
					},
				},
			},
		},
	}

	// Convert request to JSON
	openAIRequestJSON, err := json.Marshal(openAIRequest)
	if err != nil {
		utils.Error("Failed to marshal OpenAI request: %v", err)
		models.WriteServiceError(w, "Failed to prepare API request", false, true, http.StatusInternalServerError)
		return
	}

	// Get OpenAI API key from environment variable or configuration
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		utils.Error("OpenAI API key not configured on server")
		models.WriteServiceError(w, "Server configuration error", false, true, http.StatusInternalServerError)
		return
	}

	// Make request to OpenAI API
	utils.Info("Sending request to OpenAI API")
	client := &http.Client{}
	openAIReq, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(openAIRequestJSON))
	if err != nil {
		utils.Error("Failed to create OpenAI request: %v", err)
		models.WriteServiceError(w, "Failed to create API request", false, true, http.StatusInternalServerError)
		return
	}

	openAIReq.Header.Set("Content-Type", "application/json")
	openAIReq.Header.Set("Authorization", apiKey)

	response, err := client.Do(openAIReq)
	if err != nil {
		utils.Error("Failed to send request to OpenAI: %v", err)
		models.WriteServiceError(w, "Failed to connect to OpenAI API", false, true, http.StatusServiceUnavailable)
		return
	}
	defer response.Body.Close()

	// Read OpenAI response
	utils.Debug("Reading OpenAI API response")
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		utils.Error("Failed to read OpenAI response: %v", err)
		models.WriteServiceError(w, "Failed to read API response", false, true, http.StatusInternalServerError)
		return
	}

	// Check for non-200 response
	if response.StatusCode != http.StatusOK {
		utils.Error("OpenAI API returned non-200 status code: %d, response: %s", response.StatusCode, string(responseBody))
		models.WriteServiceError(w, fmt.Sprintf("OpenAI API error: %s", string(responseBody)), false, true, http.StatusInternalServerError)
		return
	}

	// Parse OpenAI response
	utils.Debug("Parsing OpenAI API response")
	var openAIResponse OpenAIResponse
	err = json.Unmarshal(responseBody, &openAIResponse)
	if err != nil {
		utils.Error("Failed to parse OpenAI response: %v", err)
		models.WriteServiceError(w, "Failed to parse API response", false, true, http.StatusInternalServerError)
		return
	}

	// Extract the analysis content
	if len(openAIResponse.Choices) == 0 {
		utils.Error("No content in OpenAI response")
		models.WriteServiceError(w, "No content returned from analysis", false, true, http.StatusInternalServerError)
		return
	}

	analysisContent := openAIResponse.Choices[0].Message.Content

	// Prepare the response
	utils.Info("Preparing product analysis response")
	var res = map[string]interface{}{
		"analysis": analysisContent,
	}

	// Return success response
	models.WriteServiceResponse(w, "Product analysis completed", res, true, true, http.StatusOK)
}
