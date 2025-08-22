package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
	"github.com/jimyeongjung/owlverload_api/utils"
)

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
	Barcode      string    `json:"barcode"`
	Code         string    `json:"code"`
	ItemID       string    `json:"item_id"`
	StockType    StockType `json:"stock_type"`
	Quantity     int       `json:"quantity"`
	ExpiryDate   time.Time `json:"expiry_date"`
	Location     string    `json:"location"`
	UserID       string    `json:"user_id"`
	Notes        string    `json:"notes"`
	DiscountRate int       `json:"discount_rate"`
}

type StockOutRequest struct {
	Stock     models.Stock `json:"stock"`
	StockType StockType    `json:"stock_type"` // BOX, BUNDLE, SINGLE
	Quantity  int          `json:"quantity"`
	UserEmail string       `json:"user_email"`
	Date      time.Time    `json:"date"`
	Notes     string       `json:"notes"`
}
type StockType string

const (
	StockTypeBox    StockType = "BOX"
	StockTypeBundle StockType = "BUNDLE"
	StockTypePCS    StockType = "PCS"
)

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
	SearchType string `json:"search_type"` // item_code | barcode | name
	Value      string `json:"value"`
}

func HandleGetItemById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- HandleGetItemById started --- ")
	// /api/v1/getItemById?itemId=${id}
	itemId := r.URL.Query().Get("itemId")

	item, err := models.GetItemById(itemId)
	if err != nil {
		models.WriteServiceError(w, "Item not found", false, true, http.StatusNotFound)
		return
	}
	stocks, err := models.GetStocksByItemId(itemId)
	if err != nil {
		models.WriteServiceError(w, "Stocks not found", false, true, http.StatusNotFound)
		return
	}
	item.Stock = stocks
	models.WriteServiceResponse(w, "Item found", item, true, true, http.StatusOK)
	fmt.Println("--- HandleGetItemById ended --- ")
}

func HandleUpdateItemById(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- HandleUpdateItemById started --- ")
	itemId := r.URL.Query().Get("itemId")
	item, err := models.GetItemById(itemId)
	if err != nil {
		models.WriteServiceError(w, "Item not found", false, true, http.StatusNotFound)
		return
	}
	models.WriteServiceResponse(w, "Item found", item, true, true, http.StatusOK)
	fmt.Println("--- HandleUpdateItemById ended --- ")
}

func HandleGetItemByCode(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	item, err := models.GetItemByCode(code)
	if err != nil {
		models.WriteServiceError(w, "Item not found", false, true, http.StatusNotFound)
		return
	}
	models.WriteServiceResponse(w, "Item found", item, true, true, http.StatusOK)
}

// HandleGetItemByBarcode handles GET requests to get an item by barcode
func HandleGetItemByBarcode(w http.ResponseWriter, r *http.Request) {

	// Get barcode from query params
	barcode := r.URL.Query().Get("barcode")
	// itemId := mux.Vars(r)["itemId"]

	if barcode == "" {
		models.WriteServiceError(w, "Either barcode or code or itemId is required", false, true, http.StatusBadRequest)
		return
	}

	var item models.Item
	var err error

	// First try to find by barcode
	if barcode != "" {
		item, err = models.GetItemByBarcode(barcode)
		if err != nil {
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
		fmt.Println("@@@ERR1", err)
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}
	var request StockInRequest
	err = json.Unmarshal(body, &request)
	if err != nil {
		fmt.Println("@@@ERR2", err)
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}
	// Validate request
	if request.ItemID == "" && request.Barcode == "" && request.Code == "" {
		fmt.Println("@@@ERR3", request)
		models.WriteServiceError(w, "Item ID, Barcode, or Code is required", false, true, http.StatusBadRequest)
		return
	}
	if request.Quantity <= 0 {
		fmt.Println("@@@ERR4", request)
		models.WriteServiceError(w, "Quantity must be greater than 0", false, true, http.StatusBadRequest)
		return
	}

	if request.StockType == "" {
		fmt.Println("@@@ERR5", request)
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
			fmt.Println("@@@ERR6", err)
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
		DiscountRate:      request.DiscountRate,
	}

	// Set the appropriate stock quantity based on type
	switch request.StockType {
	case StockTypeBox:
		stock.BoxNumber = request.Quantity
	case StockTypeBundle:
		stock.BundleNumber = request.Quantity
	case StockTypePCS:
		stock.PCSNumber = request.Quantity
	default:
		models.WriteServiceError(w, "Invalid stock type. Must be BOX, BUNDLE, or PCS", false, true, http.StatusBadRequest)
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
	stockQuery := "INSERT INTO stocks (fkproduct_id, box_number, pcs_number, bundle_number, expiry_date, location, registering_person, notes, discount_rate) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(stockQuery, stock.ItemId, stock.BoxNumber, stock.PCSNumber, stock.BundleNumber, stock.ExpiryDate, stock.Location, stock.RegisteringPerson, stock.Notes, stock.DiscountRate)

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

	// Fetch the updated stock list for the item
	updatedStocks, err := models.GetStocksByItemId(itemID)
	if err != nil {
		log.Printf("Error fetching updated stock list: %v", err)
		// Continue anyway - we'll just return the original stock data
		models.WriteServiceResponse(w, "Stock added successfully", stock, true, true, http.StatusOK)
		return
	}

	// Get the updated item data
	updatedItem, err := models.GetItemById(itemID)
	if err != nil {
		log.Printf("Error fetching updated item: %v", err)
		// Return just the updated stocks if we can't get the item
		models.WriteServiceResponse(w, "Stock added successfully", updatedStocks, true, true, http.StatusOK)
		return
	}

	// Update the item with the new stock data
	updatedItem.Stock = updatedStocks

	// Create a response with the updated item and stock information
	response := map[string]interface{}{
		"item":          updatedItem,
		"message":       "Stock added successfully",
		"updatedStocks": updatedStocks,
		"addedStock":    stock,
	}

	// Return success response with the updated item and stock information
	models.WriteServiceResponse(w, "Stock added successfully", response, true, true, http.StatusOK)
}

// HandleStockOut handles POST requests to remove stock with transaction support
func HandleStockOut(w http.ResponseWriter, r *http.Request) {
	// Add function tracing for debugging
	defer utils.Trace()()
	utils.Info("Starting HandleStockOut operation")

	// Get authenticated user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	utils.Info("User handling stock out: %s", userEmail)

	if userEmail == "" {
		utils.Warn("Unauthorized stock out attempt - missing user email")
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	utils.Debug("Reading request body")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		utils.Error("Failed to read request body: %v", err)
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var request StockOutRequest
	utils.Debug("Unmarshaling request JSON")
	err = json.Unmarshal(body, &request)
	utils.Debug("Request content: %+v", request)

	if err != nil {
		utils.Error("Invalid request format: %v", err)
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	utils.Debug("Validating request")
	if request.Stock.StockId == "" {
		utils.Warn("Missing stock ID in request")
		models.WriteServiceError(w, "Stock ID is required", false, true, http.StatusBadRequest)
		return
	}

	if request.Quantity <= 0 {
		utils.Warn("Invalid quantity in request: %d", request.Quantity)
		models.WriteServiceError(w, "Quantity must be greater than 0", false, true, http.StatusBadRequest)
		return
	}

	utils.Debug("Stock type: %s", request.StockType)
	if request.StockType == "" {
		utils.Warn("Missing stock type in request")
		models.WriteServiceError(w, "Stock type is required (BOX, BUNDLE, or SINGLE)", false, true, http.StatusBadRequest)
		return
	}

	// Calculate the amount of stock available
	deductedQuantity := request.Stock.BoxNumber - request.Quantity
	utils.Debug("Current stock: %d, Requested quantity: %d, Remaining: %d",
		request.Stock.BoxNumber, request.Quantity, deductedQuantity)

	if deductedQuantity < 0 {
		utils.Warn("Insufficient stock - current: %d, requested: %d",
			request.Stock.BoxNumber, request.Quantity)
		models.WriteServiceError(w, fmt.Sprintf("Stock can't be deducted more than you have. Current quantity %d, requested quantity %d",
			request.Stock.BoxNumber, request.Quantity), false, true, http.StatusBadRequest)
		return
	}

	// Get database instance and start a transaction
	utils.Info("Getting database instance and starting transaction")
	db := models.GetDBInstance(models.GetDBConfig())
	tx, err := db.Begin()
	if err != nil {
		utils.Error("Failed to start transaction: %v", err)
		models.WriteServiceError(w, "Internal server error", false, true, http.StatusInternalServerError)
		return
	}

	// Defer a rollback in case anything fails
	// If the transaction commits successfully, this rollback will be a no-op
	defer func() {
		if tx != nil {
			utils.Warn("Rolling back transaction - this is expected if there was an error")
			tx.Rollback()
		}
	}()

	// Update or remove stock within the transaction
	if deductedQuantity > 0 {
		// Update stock quantity
		utils.Info("Updating stock quantity")
		query := "UPDATE stocks SET box_number = box_number - ? WHERE stock_id = ?"
		_, err = tx.Exec(query, request.Quantity, request.Stock.StockId)
		if err != nil {
			utils.Error("Error updating stock in transaction: %v", err)
			models.WriteServiceError(w, fmt.Sprintf("Failed to update stock: %v", err), false, true, http.StatusInternalServerError)
			return
		}
		utils.Debug("Stock quantity updated successfully")
	} else if deductedQuantity == 0 {
		// Remove stock completely
		utils.Info("Removing stock completely")
		query := "DELETE FROM stocks WHERE stock_id = ?"
		_, err = tx.Exec(query, request.Stock.StockId)
		if err != nil {
			utils.Error("Error removing stock in transaction: %v", err)
			models.WriteServiceError(w, fmt.Sprintf("Failed to remove stock: %v", err), false, true, http.StatusInternalServerError)
			return
		}
		utils.Debug("Stock removed successfully")
	}

	// Record the transaction within the same DB transaction
	utils.Info("Recording stock transaction")
	transactionQuery := "INSERT INTO stock_transactions (fkitem_id, quantity, transaction_type, fkuser_email) VALUES (?, ?, ?, ?)"
	_, err = tx.Exec(transactionQuery, request.Stock.ItemId, request.Quantity, "out", userEmail)
	if err != nil {
		utils.Error("Error recording transaction in DB: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to record transaction: %v", err), false, true, http.StatusInternalServerError)
		return
	}
	utils.Debug("Transaction recorded successfully")

	// Commit the transaction
	utils.Info("Committing transaction")
	err = tx.Commit()
	if err != nil {
		utils.Error("Error committing transaction: %v", err)
		models.WriteServiceError(w, "Failed to complete the stock operation. Please try again.", false, true, http.StatusInternalServerError)
		return
	}
	utils.Info("Transaction committed successfully")

	// Set tx to nil to prevent the deferred rollback from doing anything
	tx = nil

	// Fetch the updated stock list for the item
	utils.Info("Fetching updated stock list for item: %s", request.Stock.ItemId)
	updatedStocks, err := models.GetStocksByItemId(request.Stock.ItemId)
	if err != nil {
		utils.Error("Error fetching updated stock list: %v", err)
		// Continue anyway - we'll just return a success message without the updated stock list
		models.WriteServiceResponse(w, "Stock removed successfully", nil, true, true, http.StatusOK)
		return
	}
	utils.Debug("Successfully fetched %d updated stock records", len(updatedStocks))

	// Get the updated item data
	utils.Info("Fetching updated item data for item: %s", request.Stock.ItemId)
	updatedItem, err := models.GetItemById(request.Stock.ItemId)
	if err != nil {
		utils.Error("Error fetching updated item: %v", err)
		// Return just the updated stocks if we can't get the item
		models.WriteServiceResponse(w, "Stock removed successfully", updatedStocks, true, true, http.StatusOK)
		return
	}
	utils.Debug("Successfully fetched updated item data")

	// Update the item with the new stock data
	updatedItem.Stock = updatedStocks

	// Create a response with the updated item and stock information
	utils.Info("Preparing response with updated item and stock information")
	response := map[string]interface{}{
		"item":          updatedItem,
		"message":       "Stock removed successfully",
		"updatedStocks": updatedStocks,
	}

	// Return success response with the updated stock list
	utils.Info("Stock out operation completed successfully")
	models.WriteServiceResponse(w, "Stock removed successfully", response, true, true, http.StatusOK)
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
	fmt.Println("@@@BODY", string(body))
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, true, http.StatusBadRequest)
		return
	}

	var item models.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		fmt.Println("@@@err", err)
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// // Validate the item
	// if item.BarCode == "" {
	// 	models.WriteServiceError(w, "Barcode is required", false, true, http.StatusBadRequest)
	// 	return
	// }

	if item.Name == "" {
		models.WriteServiceError(w, "Name is required", false, true, http.StatusBadRequest)
		return
	}

	if item.Code == "" {
		models.WriteServiceError(w, "Code is required", false, true, http.StatusBadRequest)
		return
	}

	// // Check if item with the same barcode or code already exists
	// existingByBarcode, err := models.GetItemByBarcode(item.BarCode)
	// if err == nil && existingByBarcode.ID != "" {
	// 	models.WriteServiceError(w, "An item with this barcode already exists", false, true, http.StatusBadRequest)
	// 	return
	// }

	existingByCode, err := models.GetItemByCode(item.Code)
	if err == nil && existingByCode.ID != "" {
		models.WriteServiceError(w, "An item with this code already exists", false, true, http.StatusBadRequest)
		return
	}

	// Generate ID if not provided
	if item.ID == "" {
		item.ID = fmt.Sprintf("item_%d", time.Now().UnixNano())
	}

	// fix file name
	filename, err := extractFilenameFromPath(item.ImagePath)
	filename = "/" + filename
	item.ImagePath = filename

	if err != nil {
		models.WriteServiceError(w, "Invalid image path format", false, true, http.StatusBadRequest)
		return
	}
	item.ImagePath = filename

	if item.Price == 0 {
		item.Price = 0
	}

	if item.BoxPrice == 0 {
		item.BoxPrice = 0
	}

	// Create the item (this will also handle tag associations)
	createdItem, err := models.CreateItem(item)
	if err != nil {
		log.Printf("Error creating item: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to create item: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// Fetch the complete item with all associated data for the response
	completeItem, err := models.GetItemById(createdItem.ID)
	if err != nil {
		log.Printf("Error fetching complete item data: %v", err)
		// Continue with the basic item data if we can't fetch complete data
		completeItem = createdItem
	}

	// Get associated tags for the response
	tags, err := models.GetTagsForItem(completeItem.ID)
	if err != nil {
		log.Printf("Error fetching tags for item: %v", err)
		// Continue with empty tags if error
		tags = []models.Tag{}
	}
	completeItem.Tag = tags

	// Get stocks for the item (should be empty for new items)
	stocks, err := models.GetStocksByItemId(completeItem.ID)
	if err != nil {
		log.Printf("Error fetching stocks for item: %v", err)
		stocks = []models.Stock{} // Empty array if error
	}
	completeItem.Stock = stocks

	// Extract tag names for the response
	var tagNames []string
	for _, tag := range completeItem.Tag {
		tagNames = append(tagNames, tag.TagName)
	}

	// Prepare response with item and tag information
	response := map[string]interface{}{
		"item":      completeItem,
		"tag_names": tagNames,
		"message":   "Item created successfully",
	}

	models.WriteServiceResponse(w, "Item created successfully", response, true, true, http.StatusOK)
}

// HandleRegisterItem is an alias for HandleCreateItem for API consistency
func HandleRegisterItem(w http.ResponseWriter, r *http.Request) {
	// Simply call HandleCreateItem - this is just an alias for API consistency
	HandleCreateItem(w, r)
}

// HandleUpdateItem handles PUT requests to update an existing item
func HandleUpdateItem(w http.ResponseWriter, r *http.Request) {
	fmt.Println("---HANDLEUPDATEITEM---")

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
	fmt.Println("@@@item@@@@@@@@@@@@@@@@@@@@@ passed receved", string(body))

	var item models.Item
	err = json.Unmarshal(body, &item)
	if err != nil {
		fmt.Println("@@@err@@@@@@@@@@@@@@@@@@@@@ passed unmarshal", err)
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}
	fmt.Println("@@@item@@@@@@@@@@@@@@@@@@@@@ passed unmarshal", item)

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
		"item":      updatedItem,
		"tag_names": tagNames,
		"message":   "Item updated successfully",
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
			"item":      item,
			"tag_names": tagNames,
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
			"item":      item,
			"tag_names": tagNames,
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
		fmt.Println("@@@err1", err)
		models.WriteServiceError(w, "Invalid request format", false, true, http.StatusBadRequest)
		return
	}

	// Validate request
	if lookupRequest.SearchType == "" {
		fmt.Println("@@@err2", err)
		models.WriteServiceError(w, "Search type is required (code, barcode, or name)", false, true, http.StatusBadRequest)
		return
	}

	if lookupRequest.Value == "" {
		fmt.Println("@@@err3", err)
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
	// fmt.Println("@@items", items)
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
		searchItemData := map[string]interface{}{
			"item":      item,
			"tag_names": tagNames,
		}

		itemsWithStockAndTags = append(itemsWithStockAndTags, searchItemData)
	}

	// Prepare the response
	response := map[string]interface{}{
		"searchItems": itemsWithStockAndTags,
		"total":       len(items),
		"searchType":  lookupRequest.SearchType,
		"searchValue": lookupRequest.Value,
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

// HandleGetItemsPaginated handles GET requests to get items with pagination
func HandleGetItemsPaginated(w http.ResponseWriter, r *http.Request) {
	// Get authentication user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Get pagination parameters from query string
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")
	tagParams := r.URL.Query()["tag"]

	if len(tagParams) == 0 {
		tagParams = []string{"sp6"}
	}
	// Set default values
	page := 1
	limit := 10

	// Parse page parameter
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	// Parse limit parameter
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Fetch paginated items from database
	items, totalCount, err := models.GetItemsPaginated(offset, limit, tagParams)
	if err != nil {
		log.Printf("Error retrieving paginated items: %v", err)
		models.WriteServiceError(w, "Failed to retrieve items", false, true, http.StatusInternalServerError)
		return
	}

	// For each item, populate stock and tag information
	// var itemsWithStockAndTags []map[string]interface{}
	// for _, item := range items {

	// Extract tag names for simplicity in the response
	// var tagNames []string
	// items[i].Tag = tagNames
	// for _, tag := range item.Tag {
	// 	tagNames = append(tagNames, tag.TagName)
	// }

	// itemData := map[string]interface{}{
	// 	"products": item,
	// 	"tags":     tagNames,
	// }

	// itemsWithStockAndTags = append(itemsWithStockAndTags, itemData)
	// }

	// Calculate pagination metadata
	totalPages := (totalCount + limit - 1) / limit
	hasNext := page < totalPages
	hasPrev := page > 1

	// Prepare the response with pagination metadata
	response := map[string]interface{}{
		"items": items,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total_items": totalCount,
			"total_pages": totalPages,
			"has_next":    hasNext,
			"has_prev":    hasPrev,
		},
	}

	models.WriteServiceResponse(w, "Items retrieved successfully", response, true, true, http.StatusOK)
}

// HandleGetItemsExpiringWithinDays handles GET requests to get items expiring within specific days
func HandleGetItemsExpiringWithinDays(w http.ResponseWriter, r *http.Request) {
	// Get authentication user ID from context
	tokenClaims := firebase.GetTokenClaimsFromContext(r.Context())
	userEmail := tokenClaims.Email
	fmt.Println("@@@USER NAME", userEmail)
	if userEmail == "" {
		models.WriteServiceError(w, "User authentication required", false, true, http.StatusUnauthorized)
		return
	}

	// Get 'within' parameter from query string
	withinStr := r.URL.Query().Get("within")
	if withinStr == "" {
		models.WriteServiceError(w, "Parameter 'within' is required (number of days)", false, true, http.StatusBadRequest)
		return
	}

	// Parse the within parameter
	withinDays, err := strconv.Atoi(withinStr)
	if err != nil || withinDays < 0 {
		models.WriteServiceError(w, "Parameter 'within' must be a positive integer representing days", false, true, http.StatusBadRequest)
		return
	}

	// Validate within days range (optional - set reasonable limits)
	if withinDays > 365 {
		models.WriteServiceError(w, "Parameter 'within' cannot exceed 365 days", false, true, http.StatusBadRequest)
		return
	}

	// Get items expiring within the specified days
	expiringItems, err := models.GetItemsExpiringWithinDays(withinDays)
	if err != nil {
		log.Printf("Error retrieving expiring items: %v", err)
		models.WriteServiceError(w, fmt.Sprintf("Failed to retrieve expiring items: %v", err), false, true, http.StatusInternalServerError)
		return
	}

	// If no items found, return empty array but with success status
	if len(expiringItems) == 0 {
		response := map[string]interface{}{
			"expiringItems": []models.ItemWithDaysToExpiry{},
			"total":         0,
			"withinDays":    withinDays,
			"message":       fmt.Sprintf("No items found expiring within %d days", withinDays),
		}
		models.WriteServiceResponse(w, fmt.Sprintf("No items expiring within %d days", withinDays), response, true, true, http.StatusOK)
		return
	}

	// Process the results to add tag names for convenience
	var enrichedResults []map[string]interface{}
	for _, expiringItem := range expiringItems {
		// Extract tag names for simplicity in the response
		var tagNames []string
		for _, tag := range expiringItem.Item.Tag {
			tagNames = append(tagNames, tag.TagName)
		}

		// Create enriched result with additional metadata
		enrichedResult := map[string]interface{}{
			"item":           expiringItem.Item,
			"days_to_expiry": expiringItem.DaysToExpiry,
			"stock_id":       expiringItem.StockId,
			"tag_names":      tagNames,
		}

		enrichedResults = append(enrichedResults, enrichedResult)
	}

	// Prepare the response with metadata
	response := map[string]interface{}{
		"expiring_items": enrichedResults,
		"total":          len(expiringItems),
		"within_days":    withinDays,
		"message":        fmt.Sprintf("Found %d items expiring within %d days", len(expiringItems), withinDays),
	}

	models.WriteServiceResponse(w, fmt.Sprintf("Found %d items expiring within %d days", len(expiringItems), withinDays), response, true, true, http.StatusOK)
}
