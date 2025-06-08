package apis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
	"github.com/jimyeongjung/owlverload_api/utils"
)

// BarcodeAnalyzeRequest defines the request structure for barcode analysis
type BarcodeAnalyzeRequest struct {
	Barcode     string `json:"barcode"`      // Barcode number
	ProductName string `json:"product_name"` // Product name
}

// OpenAIChatRequest defines the request structure for OpenAI API
type OpenAIChatRequest struct {
	Model    string              `json:"model"`
	Messages []OpenAIChatMessage `json:"messages"`
}

type OpenAIChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIChatResponse defines the response structure from OpenAI API
type OpenAIChatResponse struct {
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

// ProductAnalysisResult defines the structured analysis result from OpenAI
type ProductAnalysisResult struct {
	Name struct {
		English  string `json:"english"`
		Korean   string `json:"korean"`
		Japanese string `json:"japanese"`
		Chinese  string `json:"chinese"`
	} `json:"name"`
	ExpiryDate            string `json:"expiry_date"`
	IngredientsTranslated string `json:"ingredients_translated"`
	ContainsAlcohol       string `json:"contains_alcohol"`
	HalalStatus           string `json:"halal_status"`
	ContainsPork          string `json:"contains_pork"`
	ContainsBeef          string `json:"contains_beef"`
	IsPlantBased          string `json:"is_plant_based"`
	Reasoning             string `json:"reasoning"`
}

// HandleBarcodeAnalyze handles the request to analyze a barcode
func HandleBarcodeAnalyze(w http.ResponseWriter, r *http.Request) {
	defer utils.Trace()()
	utils.Info("Starting HandleBarcodeAnalyze operation")

	// Get authenticated user from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	utils.Info("User handling barcode analysis: %s", userEmail)

	if userEmail == "" {
		utils.Warn("Unauthorized barcode analysis attempt - missing user email")
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

	var request BarcodeAnalyzeRequest
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
	fmt.Println("request.ProductName", request.ProductName)

	// Prepare OpenAI request
	utils.Debug("Preparing OpenAI request")
	openAIRequest := OpenAIChatRequest{
		Model: "gpt-4o",
		Messages: []OpenAIChatMessage{
			{
				Role:    "system",
				Content: "You are a product analysis assistant. For the given barcode, provide information about the product in JSON format. If you don't know about a specific barcode, provide a response indicating that the product information is not available.",
			},
			{
				Role: "user",
				Content: fmt.Sprintf(`Analyze korean noodles, name %s with barcode: %s. Return the information in this exact JSON format:

				{
				"name": {
					"english": "",
					"korean": "",
					"japanese": "",
					"chinese": ""
				},
				"ingredients_translated": "",
				"contains_alcohol": "Yes" or "No" or "Unclear",
				"contains_pork": "Yes" or "No" or "Unclear",
				"contains_beef": "Yes" or "No" or "Unclear",
				"is_plant_based": "Yes" or "No" or "Unclear",
				"halal_status": "Halal" or "Not Halal" or "Unclear",
				"reasoning": ""
				}
				If you don't know this product, provide best guesses and indicate uncertainty in the reasoning field.
				Do not include any other text in your response.`, request.ProductName, request.Barcode),
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

	// Get OpenAI API key from environment variable
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
	openAIReq.Header.Set("Authorization", "Bearer "+apiKey)

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
	var openAIResponse OpenAIChatResponse
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

	// Parse the JSON response from GPT
	utils.Debug("Parsing analysis content into structured format")
	var analysisResult ProductAnalysisResult
	err = json.Unmarshal([]byte(analysisContent), &analysisResult)
	if err != nil {
		utils.Error("Failed to parse analysis content as JSON: %v", err)
		models.WriteServiceError(w, "Failed to parse analysis content", false, true, http.StatusInternalServerError)
		return
	}

	// Create or update item in the database
	var item models.Item

	// Update item with analysis data
	item.Name = analysisResult.Name.English
	item.NameEng = analysisResult.Name.English
	item.NameKor = analysisResult.Name.Korean
	item.NameJpn = analysisResult.Name.Japanese
	item.NameChn = analysisResult.Name.Chinese
	item.Ingredients = analysisResult.IngredientsTranslated
	item.IsHalal = analysisResult.HalalStatus == "Halal"
	item.Reasoning = analysisResult.Reasoning

	// Set pork, beef, plant-based fields if present in model
	if analysisResult.ContainsPork != "" {
		if analysisResult.ContainsPork == "Yes" {
			item.IsPorkContained = true
		} else if analysisResult.ContainsPork == "No" {
			item.IsPorkContained = false
		} else {
			// "Unclear" or unknown: leave as-is or set to nil/false
		}
	}
	if analysisResult.ContainsBeef != "" {
		if analysisResult.ContainsBeef == "Yes" {
			item.IsBeefContained = true
		} else if analysisResult.ContainsBeef == "No" {
			item.IsBeefContained = false
		} else {
			// "Unclear" or unknown: leave as-is or set to nil/false
		}
	}
	if analysisResult.IsPlantBased != "" {
		if analysisResult.IsPlantBased == "Yes" {
			item.IsPlantBased = true
		} else if analysisResult.IsPlantBased == "No" {
			item.IsPlantBased = false
		} else {
			// "Unclear" or unknown: leave as-is or set to nil/false
		}
	}

	// Prepare the response
	utils.Info("Preparing barcode analysis response")
	analysisResponse := map[string]interface{}{
		"analysis": analysisResult,
	}

	// Return success response
	models.WriteServiceResponse(w, "Barcode analysis completed", analysisResponse, true, true, http.StatusOK)
}

// "The product information for barcode 8801043060554 is not available in the current database. Without specific details about the product, uncertainty exists regarding its ingredients and dietary status."
