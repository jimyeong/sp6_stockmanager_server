package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
)

// Request structures for tag operations
type GetRecommendationsRequest struct {
	TagIDs []string `json:"tagIds"`
	Limit  int      `json:"limit"`
	Page   int      `json:"page"`
}

type CreateTagRequest struct {
	TagName string `json:"tag_name"`
}

type AssociateTagsRequest struct {
	ItemID string   `json:"itemId"`
	TagIDs []string `json:"tag_ids"`
}

// HandleGetRecommendedItems handles requests for item recommendations based on tags with pagination
func HandleGetRecommendedItems(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var recommendRequest GetRecommendationsRequest
	err = json.Unmarshal(body, &recommendRequest)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	if len(recommendRequest.TagIDs) == 0 {
		models.WriteServiceError(w, "At least one tag ID is required", false, true, http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if recommendRequest.Limit <= 0 {
		recommendRequest.Limit = 10
	}

	if recommendRequest.Page <= 0 {
		recommendRequest.Page = 1
	}

	// Get recommended items based on tags with pagination
	items, totalCount, err := models.GetRecommendedItems(
		recommendRequest.TagIDs,
		recommendRequest.Limit,
		recommendRequest.Page,
	)

	if err != nil {
		log.Printf("Error getting recommended items: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to get recommendations: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// If no items found, return an empty array but with success status
	if len(items) == 0 {
		// Prepare empty response with pagination info
		emptyResponse := map[string]interface{}{
			"itemDetails": []models.Item{},
			"total":       totalCount,
			"tagIds":      recommendRequest.TagIDs,
			"requested":   recommendRequest.Limit,
			"page":        recommendRequest.Page,
		}

		models.WriteServiceResponse(w, "No recommendations found for the provided tags", emptyResponse, true, true, http.StatusOK)
		return
	}

	// Extract tag names for each item for easier access on the client side
	// var itemsWithTagNames []map[string]interface{}
	var itemsWithTagNames []models.Item
	itemsWithTagNames = items
	// for _, item := range items {
	// 	// var tagNames []string
	// 	// for _, tag := range item.Tag {
	// 	// 	tagNames = append(tagNames, tag.TagName)
	// 	// }

	// 	// itemData := map[string]interface{}{
	// 	// 	"tagNames": tagNames,
	// 	// 	"item":     item,
	// 	// }

	// 	itemsWithTagNames = append(itemsWithTagNames, item)
	// }

	// Prepare the response with pagination info
	payload := map[string]interface{}{
		"itemDetails": itemsWithTagNames,
		"total":       totalCount,
		"tagIds":      recommendRequest.TagIDs,
		"requested":   recommendRequest.Limit,
		"page":        recommendRequest.Page,
	}

	// Format response according to required structure
	response := map[string]interface{}{
		"status":     http.StatusOK,
		"payload":    payload,
		"userExists": true,
		"success":    true,
	}

	// Use direct JSON encoding to match the exact format requested
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleGetAllTags handles requests to get all tags
func HandleGetAllTags(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Check for optional category filter
	// category := r.URL.Query().Get("category")

	var tags []models.Tag
	var err error

	// Get tags, either all or filtered by category
	// if category != "" {
	// 	tags, err = models.GetTagsByCategory(category)
	// } else {

	// }
	tags, err = models.GetAllTags()
	fmt.Println(tags)

	if err != nil {
		log.Printf("Error retrieving tags: %v", err)
		models.WriteServiceError(w, "Failed to retrieve tags", false, true, http.StatusInternalServerError)
		return
	}

	payload := map[string]interface{}{
		"tags": tags,
	}
	// If no tags found, return an empty array but with success status
	if len(tags) == 0 {
		models.WriteServiceResponse(w, "No tags found", []models.Tag{}, true, true, http.StatusOK)
		return
	}

	models.WriteServiceResponse(w, "Tags retrieved successfully", payload, true, true, http.StatusOK)
}

// HandleCreateTag handles requests to create a new tag
func HandleCreateTag(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var tagRequest CreateTagRequest
	err = json.Unmarshal(body, &tagRequest)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	if tagRequest.TagName == "" {
		models.WriteServiceError(w, "Tag name is required", false, true, http.StatusBadRequest)
		return
	}

	// Create the tag
	tag := models.Tag{
		TagName: tagRequest.TagName,
	}

	createdTag, err := models.CreateTag(tag)
	if err != nil {
		log.Printf("Error creating tag: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to create tag: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	models.WriteServiceResponse(w, "Tag created successfully", createdTag, true, true, http.StatusCreated)
}

// HandleAssociateItemWithTags handles requests to associate an item with tags
func HandleAssociateItemWithTags(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	userID := tokenClaims.UID
	fmt.Println("@@@USER NAME", userEmail)
	if userID == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var associateRequest AssociateTagsRequest
	err = json.Unmarshal(body, &associateRequest)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	if associateRequest.ItemID == "" {
		models.WriteServiceError(w, "Item ID is required", false, true, http.StatusBadRequest)
		return
	}

	if len(associateRequest.TagIDs) == 0 {
		models.WriteServiceError(w, "At least one tag ID is required", false, true, http.StatusBadRequest)
		return
	}

	// Associate the item with the tags
	err = models.AssociateItemWithTags(associateRequest.ItemID, associateRequest.TagIDs)
	if err != nil {
		log.Printf("Error associating item with tags: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to associate item with tags: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// Get all tags for the item after association
	tags, err := models.GetTagsForItem(associateRequest.ItemID)
	if err != nil {
		log.Printf("Error getting tags for item: %v", err)
		// Continue anyway and just return success
		models.WriteServiceResponse(w, "Item associated with tags successfully", nil, true, true, http.StatusOK)
		return
	}

	models.WriteServiceResponse(w, "Item associated with tags successfully", tags, true, true, http.StatusOK)
}

// HandleGetTagsForItem handles requests to get all tags for a specific item
func HandleGetTagsForItem(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	userID := tokenClaims.UID
	fmt.Println("@@@USER NAME", userEmail)
	if userID == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Get item ID from query params
	itemID := mux.Vars(r)["itemId"]
	if itemID == "" {
		models.WriteServiceError(w, "Item ID is required", false, true, http.StatusBadRequest)
		return
	}

	// Get tags for the item
	tags, err := models.GetTagsForItem(itemID)
	if err != nil {
		log.Printf("Error getting tags for item: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to get tags for item: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// If no tags found, return an empty array but with success status
	if len(tags) == 0 {
		models.WriteServiceResponse(w, "No tags found for the item", []models.Tag{}, true, true, http.StatusOK)
		return
	}

	models.WriteServiceResponse(w, "Tags retrieved successfully", tags, true, true, http.StatusOK)
}

// HandleGetPopularTags handles requests to get the most popular tags
func HandleGetPopularTags(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	userID := tokenClaims.UID
	fmt.Println("@@@USER NAME", userEmail)
	if userID == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Get limit from query params
	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Default limit
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit <= 0 {
			models.WriteServiceError(w, "Invalid limit parameter", false, true, http.StatusBadRequest)
			return
		}
	}

	// Get popular tags
	tags, err := models.GetPopularTags(limit)
	if err != nil {
		log.Printf("Error getting popular tags: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to get popular tags: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// If no tags found, return an empty array but with success status
	if len(tags) == 0 {
		models.WriteServiceResponse(w, "No tags found", []models.Tag{}, true, true, http.StatusOK)
		return
	}

	models.WriteServiceResponse(w, "Popular tags retrieved successfully", tags, true, true, http.StatusOK)
}

// HandleSearchTags handles requests to search for tags by name
func HandleSearchTags(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Get search term from query params
	searchTerm := r.URL.Query().Get("q")
	if searchTerm == "" {
		models.WriteServiceError(w, "Search term is required", false, true, http.StatusBadRequest)
		return
	}

	// Search for tags
	tags, err := models.SearchTagsByName(searchTerm)
	if err != nil {
		log.Printf("Error searching for tags: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to search for tags: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// If no tags found, return an empty array but with success status
	if len(tags) == 0 {
		models.WriteServiceResponse(w, "No tags found matching the search term", []models.Tag{}, true, true, http.StatusOK)
		return
	}

	models.WriteServiceResponse(w, "Tags found successfully", tags, true, true, http.StatusOK)
}
