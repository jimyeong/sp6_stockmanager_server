package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
)

type ItemStock struct {
	stockId           string
	itemId            string
	boxNumber         string
	bundleNumber      string
	singleNumber      string
	expiryDate        time.Time
	registeredDate    time.Time
	createdAt         time.Time
	location          string
	registeringPerson string
	notes             string
}
type Item struct {
	id                string      `json:"id"`
	code              string      `json:"code"`
	barcode           string      `json:"barcode"`
	name              string      `json:"name"`
	itemType          string      `json:"itemType"`
	availableForOrder bool        `json:"availableForOrder"`
	imagePath         string      `json:"imagePath"`
	createdAt         time.Time   `json:"createdAt"`
	stock             []ItemStock `json:"stock"`
}

// Request structures
type GetItemRequest struct {
	Barcode string `json:"barcode"`
	Code    string `json:"code"`
}

// id: string;
// type: 'IN' | 'OUT';
// stockType: 'BOX' | 'BUNDLE' | 'SINGLE';
// quantity: number;
// date: Date;
// userId: string;
type StockInRequest struct {
	Barcode    string    `json:"barcode"`
	Code       string    `json:"code"`
	ItemID     string    `json:"itemId"`
	StockType  string    `json:"stockType"`
	Quantity   int       `json:"quantity"`
	ExpiryDate time.Time `json:"expiryDate"`
	Location   string    `json:"location"`
	UserID     string    `json:"userId"`
	Notes      string    `json:"notes"`
}

type StockOutRequest struct {
	Stock     models.Stock `json:"stock"`
	StockType string       `json:"stockType"` // BOX, BUNDLE, SINGLE
	Quantity  int          `json:"quantity"`
	UserEmail string       `json:"userEmail"`
	Date      time.Time    `json:"date"`
	Notes     string       `json:"notes"`
}

// SearchItemsRequest defines parameters for searching items
type SearchItemsRequest struct {
	Query    string `json:"query"`
	Category string `json:"category"`
	Type     string `json:"type"`
	Limit    int    `json:"limit"`
	Offset   int    `json:"offset"`
}

// LookupItemRequest defines parameters for looking up items by field
type LookupItemRequest struct {
	SearchType string `json:"search_type"` // code, barcode, or name
	Value      string `json:"value"`
}

// HandleGetItemByBarcode handles GET requests to get an item by barcode
func HandleGetItemByBarcode(w http.ResponseWriter, r *http.Request) {
	// Get barcode from query params
	barcode := r.URL.Query().Get("barcode")
	code := r.URL.Query().Get("code")
	fmt.Println("barcode", barcode)

	if barcode == "" && code == "" {
		models.WriteServiceError(w, "Either barcode or code is required", false, true, http.StatusBadRequest)
		return
	}

	var item models.Item
	var err error

	// First try to find by barcode
	if barcode != "" {
		item, err = models.GetItemByBarcode(barcode)
		fmt.Println("--- item, err --- ", item, err)
		fmt.Println("--- err --- ", err)
		if err != nil {
			// induce themto send createItem request
			// models.WriteServiceError(w, "Item not found", false, true)
			// w.WriteHeader(http.StatusNotFound)
			w.WriteHeader(204)
			json.NewEncoder(w).Encode(models.ServiceResponse{
				Message: "Item not found",
				Payload: map[string]interface{}{
					"item":    item,
					"barcode": barcode,
				},
				Success:    true,
				UserExists: true,
			})
			return
			// return
		}
		models.WriteServiceResponse(w, "Item found", item, true, true, http.StatusOK)
		return
	}

	// If barcode search failed or wasn't provided, try by code
	if code != "" {
		// Assuming we have a GetItemByCode function
		item, err = models.GetItemByCode(code)
		if err == nil {
			models.WriteServiceResponse(w, "Item found", item, true, true, http.StatusOK)
			return
		}
	}

	// If we get here, item wasn't found
	log.Printf("Error getting item: %v", err)
	models.WriteServiceError(w, "Item not found", true, true, http.StatusNotFound)
}

// HandleStockIn handles POST requests to add stock with transaction support
func HandleStockIn(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userName := tokenClaims.DisplayName
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userName)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var request StockInRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	if request.ItemID == "" && request.Barcode == "" && request.Code == "" {
		models.WriteServiceError(w, "Item ID, Barcode, or Code is required", false, true, http.StatusBadRequest)
		return
	}

	if request.Quantity <= 0 {
		models.WriteServiceError(w, "Quantity must be greater than 0", false, true, http.StatusBadRequest)
		return
	}

	if request.StockType == "" {
		models.WriteServiceError(w, "Stock type is required (BOX, BUNDLE, or SINGLE)", false, true, http.StatusBadRequest)
		return
	}

	// If itemID is not provided, try to get it from barcode or code
	var itemID string
	if request.ItemID != "" {
		itemID = request.ItemID
	} else {
		var item models.Item
		if request.Barcode != "" {
			item, err = models.GetItemByBarcode(request.Barcode)
		} else if request.Code != "" {
			item, err = models.GetItemByCode(request.Code)
		}

		if err != nil {
			models.WriteServiceError(w, fmt.Sprintf("Failed to find item: %v", err), false, true, http.StatusNotFound)
			return
		}
		itemID = item.ID
	}

	// Create a new stock record
	stock := models.Stock{
		StockId:           fmt.Sprintf("stock_%d", time.Now().UnixNano()),
		ItemId:            itemID,
		ExpiryDate:        request.ExpiryDate,
		Notes:             request.Notes,
		CreatedAt:         time.Now(),
		Location:          request.Location,
		RegisteringPerson: userName,
	}

	// Set the appropriate stock quantity based on type
	switch request.StockType {
	case "BOX":
		stock.BoxNumber = request.Quantity
	case "BUNDLE":
		stock.BundleNumber = request.Quantity
	case "SINGLE":
		stock.SingleNumber = request.Quantity
	default:
		models.WriteServiceError(w, "Invalid stock type. Must be BOX, BUNDLE, or SINGLE", false, true, http.StatusBadRequest)
		return
	}

	// Get database instance and start a transaction
	db := models.GetDBInstance(models.GetDBConfig())
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		models.WriteServiceError(w, "Internal server error", false, true, http.StatusInternalServerError)
		return
	}

	// Defer a rollback in case anything fails
	// If the transaction commits successfully, this rollback will be a no-op
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// Insert stock within transaction
	stockQuery := "INSERT INTO stocks (fkproduct_id, box_number, single_number, bundle_number, expiry_date, location, registering_person, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(stockQuery, stock.ItemId, stock.BoxNumber, stock.SingleNumber, stock.BundleNumber, stock.ExpiryDate, stock.Location, stock.RegisteringPerson, stock.Notes)
	if err != nil {
		log.Printf("Error adding stock in transaction: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to add stock: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// Record the transaction within the same DB transaction
	transactionQuery := "INSERT INTO stock_transactions (fkitem_id, quantity, transaction_type, fkuser_email) VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(transactionQuery, itemID, request.Quantity, "in", userEmail)
	if err != nil {
		log.Printf("Error recording transaction in DB transaction: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to record transaction: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		models.WriteServiceError(w, "Failed to complete the stock operation. Please try again.", false, true, http.StatusInternalServerError)
		return
	}

	// Set tx to nil to prevent the deferred rollback from doing anything
	tx = nil

	// Return success response
	models.WriteServiceResponse(w, "Stock added successfully", stock, true, true, http.StatusOK)
}

// HandleStockOut handles POST requests to remove stock with transaction support
func HandleStockOut(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var request StockOutRequest

	err = json.Unmarshal(body, &request)
	fmt.Println("---request---", request)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Stock.StockId == "" {
		models.WriteServiceError(w, "Stock ID is required", false, true, http.StatusBadRequest)
		return
	}

	if request.Quantity <= 0 {
		models.WriteServiceError(w, "Quantity must be greater than 0", false, true, http.StatusBadRequest)
		return
	}

	fmt.Println("---StockTYPE---", request.StockType)
	if request.StockType == "" {
		models.WriteServiceError(w, "Stock type is required (BOX, BUNDLE, or SINGLE)", false, true, http.StatusBadRequest)
		return
	}

	// Calculate the amount of stock available
	deductedQuantity := request.Stock.BoxNumber - request.Quantity
	if deductedQuantity < 0 {
		models.WriteServiceError(w, fmt.Sprintf("Stock can't be deducted more than you have. Current quantity %d, requested quantity %d", request.Stock.BoxNumber, request.Quantity), false, true, http.StatusBadRequest)
		return
	}

	// Get database instance and start a transaction
	db := models.GetDBInstance(models.GetDBConfig())
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to start transaction: %v", err)
		models.WriteServiceError(w, "Internal server error", false, true, http.StatusInternalServerError)
		return
	}

	// Defer a rollback in case anything fails
	// If the transaction commits successfully, this rollback will be a no-op
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// Update or remove stock within the transaction
	if deductedQuantity > 0 {
		// Update stock quantity
		query := "UPDATE stocks SET box_number = box_number - ? WHERE stock_id = ?"
		_, err = tx.Exec(query, request.Quantity, request.Stock.StockId)
		if err != nil {
			log.Printf("Error updating stock in transaction: %v", err)
			models.WriteServiceError(w, fmt.Sprintf("Failed to update stock: %v", err), false, true, http.StatusInternalServerError)
			return
		}
	} else if deductedQuantity == 0 {
		// Remove stock completely
		query := "DELETE FROM stocks WHERE stock_id = ?"
		_, err = tx.Exec(query, request.Stock.StockId)
		if err != nil {
			log.Printf("Error removing stock in transaction: %v", err)
			models.WriteServiceError(w, fmt.Sprintf("Failed to remove stock: %v", err), false, true, http.StatusInternalServerError)
			return
		}
	}

	// Record the transaction within the same DB transaction
	transactionQuery := "INSERT INTO stock_transactions (fkitem_id, quantity, transaction_type, fkuser_email) VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(transactionQuery, request.Stock.ItemId, request.Quantity, "out", userEmail)
	if err != nil {
		log.Printf("Error recording transaction in DB transaction: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to record transaction: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		models.WriteServiceError(w, "Failed to complete the stock operation. Please try again.", false, true, http.StatusInternalServerError)
		return
	}

	// Set tx to nil to prevent the deferred rollback from doing anything
	tx = nil

	// Return success response
	models.WriteServiceResponse(w, "Stock removed successfully", nil, true, true, http.StatusOK)
}

// HandleCreateItem handles POST requests to create a new item
func HandleCreateItem(w http.ResponseWriter, r *http.Request) {
	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var item models.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate the item
	if item.BarCode == "" {
		models.WriteServiceError(w, "Barcode is required", false, true, http.StatusBadRequest)
		return
	}

	if item.Name == "" {
		models.WriteServiceError(w, "Name is required", false, true, http.StatusBadRequest)
		return
	}

	if item.Code == "" {
		models.WriteServiceError(w, "Code is required", false, true, http.StatusBadRequest)
		return
	}

	// Check if item with the same barcode or code already exists
	existingByBarcode, err := models.GetItemByBarcode(item.BarCode)
	if err == nil && existingByBarcode.ID != "" {
		models.WriteServiceError(w, "An item with this barcode already exists", false, true, http.StatusBadRequest)
		return
	}

	existingByCode, err := models.GetItemByCode(item.Code)
	if err == nil && existingByCode.ID != "" {
		models.WriteServiceError(w, "An item with this code already exists", false, true, http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if item.ID == "" {
		item.ID = fmt.Sprintf("item_%d", time.Now().UnixNano())
	}

	// Set creation timestamp
	item.CreatedAt = time.Now()

	// Create the item
	createdItem, err := models.CreateItem(item)
	if err != nil {
		log.Printf("Error creating item: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to create item: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	models.WriteServiceResponse(w, "Item created successfully", createdItem, true, true, http.StatusOK)
}

// HandleRegisterItem is an alias for HandleCreateItem for API consistency
func HandleRegisterItem(w http.ResponseWriter, r *http.Request) {
	// Simply call HandleCreateItem - this is just an alias for API consistency
	HandleCreateItem(w, r)
}

// HandleUpdateItem handles PUT requests to update an existing item
func HandleUpdateItem(w http.ResponseWriter, r *http.Request) {

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

	var item models.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate the item has the required fields for update (barcode or code)
	if item.BarCode == "" && item.Code == "" && item.ID == "" {
		models.WriteServiceError(w, "At least one of barcode, code, or ID is required for update", false, true, http.StatusBadRequest)
		return
	}

	// Update the item
	updatedItem, err := models.UpdateItem(item)
	if err != nil {
		log.Printf("Error updating item: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to update item: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// Get stock and tag information for the updated item
	stocks, err := models.GetStocksByItemId(updatedItem.ID)
	if err == nil {
		updatedItem.Stock = stocks
	} else {
		updatedItem.Stock = []models.Stock{} // Empty array if error
	}

	// Fetch tags for the item
	tags, err := models.GetTagsForItem(updatedItem.ID)
	if err == nil {
		updatedItem.Tag = tags
	} else {
		updatedItem.Tag = []models.Tag{} // Empty array if error
	}

	// Extract tag names for convenience
	var tagNames []string
	for _, tag := range updatedItem.Tag {
		tagNames = append(tagNames, tag.TagName)
	}

	// Prepare response
	response := map[string]interface{}{
		"item":     updatedItem,
		"tagNames": tagNames,
		"message":  "Item updated successfully",
	}

	models.WriteServiceResponse(w, "Item updated successfully", response, true, true, http.StatusOK)
}

// HandleGetItems handles GET requests to get all items
func HandleGetItems(w http.ResponseWriter, r *http.Request) {
	// Get authentication user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Check for optional filters
	category := r.URL.Query().Get("category")
	itemType := r.URL.Query().Get("type")
	searchTerm := r.URL.Query().Get("search")

	// Fetch all items
	items, err := models.GetAllItems()
	if err != nil {
		log.Printf("Error retrieving items: %v", err)
		models.WriteServiceError(w, "Failed to retrieve items", false, true, http.StatusInternalServerError)
		return
	}

	// Apply filters if provided
	var filteredItems []models.Item
	if category != "" || itemType != "" || searchTerm != "" {
		for _, item := range items {
			// Apply category filter if provided
			if category != "" && item.Type != category {
				continue
			}

			// Apply type filter if provided
			if itemType != "" && item.Type != itemType {
				continue
			}

			// Apply search filter if provided (search in code, name, and barcode)
			if searchTerm != "" {
				// Convert everything to lowercase for case-insensitive search
				lowerSearch := strings.ToLower(searchTerm)
				lowerName := strings.ToLower(item.Name)
				lowerCode := strings.ToLower(item.Code)
				lowerBarcode := strings.ToLower(item.BarCode)

				// If search term is not found in any field, skip this item
				if !strings.Contains(lowerName, lowerSearch) &&
					!strings.Contains(lowerCode, lowerSearch) &&
					!strings.Contains(lowerBarcode, lowerSearch) {
					continue
				}
			}

			// If we get here, the item passed all filters
			filteredItems = append(filteredItems, item)
		}
	} else {
		// If no filters, use all items
		filteredItems = items
	}

	// For each item, the stock and tag information should already be populated
	// from our updated GetAllItems function
	var itemsWithStockAndTags []map[string]interface{}
	for _, item := range filteredItems {
		// Create a composite response with item, its stock, and tags
		// Use the tags field that's now included in the item

		// Extract tag names for simplicity in the response
		var tagNames []string
		for _, tag := range item.Tag {
			tagNames = append(tagNames, tag.TagName)
		}

		itemData := map[string]interface{}{
			"item":     item,
			"tagNames": tagNames,
		}

		itemsWithStockAndTags = append(itemsWithStockAndTags, itemData)
	}

	models.WriteServiceResponse(w, "Items retrieved successfully", itemsWithStockAndTags, true, true, http.StatusOK)
}

// HandleSearchItems handles POST requests to search for items with more complex criteria
func HandleSearchItems(w http.ResponseWriter, r *http.Request) {
	// Get authentication user ID from context
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

	var searchRequest SearchItemsRequest
	err = json.Unmarshal(body, &searchRequest)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Set default values if not provided
	if searchRequest.Limit <= 0 {
		searchRequest.Limit = 100 // Default limit
	}
	if searchRequest.Limit > 1000 {
		searchRequest.Limit = 1000 // Maximum limit
	}

	// Perform the search
	// In a real implementation, this would be a database query with pagination
	items, err := models.GetAllItems()
	if err != nil {
		log.Printf("Error retrieving items: %v", err)
		models.WriteServiceError(w, "Failed to retrieve items", false, true, http.StatusInternalServerError)
		return
	}

	// Filter the items based on search criteria
	var filteredItems []models.Item
	for _, item := range items {
		// Check category filter
		if searchRequest.Category != "" && item.Type != searchRequest.Category {
			continue
		}

		// Check type filter
		if searchRequest.Type != "" && item.Type != searchRequest.Type {
			continue
		}

		// Check query string (search in all text fields)
		if searchRequest.Query != "" {
			query := strings.ToLower(searchRequest.Query)
			found := false

			// Check in various fields
			if strings.Contains(strings.ToLower(item.Name), query) ||
				strings.Contains(strings.ToLower(item.Code), query) ||
				strings.Contains(strings.ToLower(item.BarCode), query) ||
				strings.Contains(strings.ToLower(item.Type), query) {
				found = true
			}

			if !found {
				continue
			}
		}

		// Item matches all criteria
		filteredItems = append(filteredItems, item)
	}

	// Apply pagination
	startIndex := searchRequest.Offset
	endIndex := startIndex + searchRequest.Limit
	if startIndex >= len(filteredItems) {
		// Start index is beyond the available items
		filteredItems = []models.Item{}
	} else if endIndex > len(filteredItems) {
		// End index is beyond the available items, adjust it
		filteredItems = filteredItems[startIndex:]
	} else {
		// Both indices are valid
		filteredItems = filteredItems[startIndex:endIndex]
	}

	// For each item in the results, the stock and tag information should already be populated
	var itemsWithStockAndTags []map[string]interface{}
	for _, item := range filteredItems {
		// Extract tag names for simplicity in the response
		var tagNames []string
		for _, tag := range item.Tag {
			tagNames = append(tagNames, tag.TagName)
		}

		// Create a composite response with item, its stock, and tags
		itemData := map[string]interface{}{
			"item":     item,
			"tagNames": tagNames,
		}

		itemsWithStockAndTags = append(itemsWithStockAndTags, itemData)
	}

	// Prepare the response with metadata
	response := map[string]interface{}{
		"items":  itemsWithStockAndTags,
		"total":  len(filteredItems),
		"offset": searchRequest.Offset,
		"limit":  searchRequest.Limit,
	}

	models.WriteServiceResponse(w, "Search results", response, true, true, http.StatusOK)
}

// HandleLookupItems handles POST requests to search for items by field with LIKE queries
func HandleLookupItems(w http.ResponseWriter, r *http.Request) {
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

	var lookupRequest LookupItemRequest
	err = json.Unmarshal(body, &lookupRequest)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	if lookupRequest.SearchType == "" {
		models.WriteServiceError(w, "Search type is required (code, barcode, or name)", false, true, http.StatusBadRequest)
		return
	}

	if lookupRequest.Value == "" {
		models.WriteServiceError(w, "Search value is required", false, true, http.StatusBadRequest)
		return
	}

	// Validate search type
	validTypes := map[string]bool{
		"code":    true,
		"barcode": true,
		"name":    true,
	}

	if !validTypes[lookupRequest.SearchType] {
		models.WriteServiceError(w, "Invalid search type. Must be code, barcode, or name", false, true, http.StatusBadRequest)
		return
	}

	// Perform the search with LIKE query
	items, err := models.SearchItemsByField(lookupRequest.SearchType, lookupRequest.Value)
	if err != nil {
		log.Printf("Error searching for items: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to search for items: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// If no items found, return an empty array but with success status
	if len(items) == 0 {
		models.WriteServiceResponse(w, "No items found matching search criteria", []models.Item{}, true, true, http.StatusOK)
		return
	}

	// Items will already have stock and tag information from our model functions
	var itemsWithStockAndTags []map[string]interface{}
	for _, item := range items {
		// Extract tag names for simplicity in the response
		var tagNames []string
		for _, tag := range item.Tag {
			tagNames = append(tagNames, tag.TagName)
		}

		// Create a response with item and tag names
		itemData := map[string]interface{}{
			"item":     item,
			"tagNames": tagNames,
		}

		itemsWithStockAndTags = append(itemsWithStockAndTags, itemData)
	}

	// Prepare the response
	response := map[string]interface{}{
		"items":      itemsWithStockAndTags,
		"total":      len(items),
		"searchType": lookupRequest.SearchType,
		"value":      lookupRequest.Value,
	}

	models.WriteServiceResponse(w, "Items found matching search criteria", response, true, true, http.StatusOK)
}

// HandleGetItemsWithMissingInfo handles GET requests to get items with missing information
func HandleGetItemsWithMissingInfo(w http.ResponseWriter, r *http.Request) {
	// Get authentication user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Fetch all items
	items, err := models.GetAllItems()
	if err != nil {
		log.Printf("Error retrieving items: %v", err)
		models.WriteServiceError(w, "Failed to retrieve items with missing information", false, true, http.StatusInternalServerError)
		return
	}

	// Filter items with missing information
	var itemsWithMissingInfo []models.Item
	var itemsMissingCode []models.Item
	var itemsMissingBarcode []models.Item
	var itemsMissingName []models.Item
	var itemsMissingImage []models.Item

	for _, item := range items {
		hasMissingInfo := false

		// Check for missing code
		if item.Code == "" {
			hasMissingInfo = true
			itemsMissingCode = append(itemsMissingCode, item)
		}

		// Check for missing barcode
		if item.BarCode == "" {
			hasMissingInfo = true
			itemsMissingBarcode = append(itemsMissingBarcode, item)
		}

		// Check for missing name
		if item.Name == "" {
			hasMissingInfo = true
			itemsMissingName = append(itemsMissingName, item)
		}

		// Check for missing image path
		if item.ImagePath == "" {
			hasMissingInfo = true
			itemsMissingImage = append(itemsMissingImage, item)
		}

		// Add item to the result list if any information is missing
		if hasMissingInfo {
			itemsWithMissingInfo = append(itemsWithMissingInfo, item)
		}
	}

	// For each item with missing info, fetch its stock information
	var itemsWithStockAndMissingInfo []map[string]interface{}
	for _, item := range itemsWithMissingInfo {
		// Get stock information for this item
		stocks, err := models.GetStocksByItemId(item.ID)
		if err != nil {
			log.Printf("Error retrieving stock for item %s: %v", item.ID, err)
			// Continue anyway, we'll just return the item without stock
			stocks = []models.Stock{} // Empty array if error
		}

		// Create a composite response with item, its stock, and missing fields
		missingFields := []string{}
		if item.Code == "" {
			missingFields = append(missingFields, "code")
		}
		if item.BarCode == "" {
			missingFields = append(missingFields, "barCode")
		}
		if item.Name == "" {
			missingFields = append(missingFields, "name")
		}
		if item.ImagePath == "" {
			missingFields = append(missingFields, "imagePath")
		}

		itemData := map[string]interface{}{
			"item":          item,
			"stock":         stocks,
			"missingFields": missingFields,
		}

		itemsWithStockAndMissingInfo = append(itemsWithStockAndMissingInfo, itemData)
	}

	// Prepare categorized response
	response := map[string]interface{}{
		"itemsWithMissingInfo": itemsWithStockAndMissingInfo,
		"missingByCategory": map[string]interface{}{
			"code":      itemsMissingCode,
			"barcode":   itemsMissingBarcode,
			"name":      itemsMissingName,
			"imagePath": itemsMissingImage,
		},
		"totalWithMissingInfo": len(itemsWithMissingInfo),
		"totalMissingCode":     len(itemsMissingCode),
		"totalMissingBarcode":  len(itemsMissingBarcode),
		"totalMissingName":     len(itemsMissingName),
		"totalMissingImage":    len(itemsMissingImage),
	}

	models.WriteServiceResponse(w, "Items with missing information retrieved successfully", response, true, true, http.StatusOK)
}
