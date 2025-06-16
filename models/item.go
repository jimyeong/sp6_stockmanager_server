package models

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jimyeongjung/owlverload_api/utils"
)

type Item struct {
	ID                string    `json:"id"`
	Code              string    `json:"code"`
	BarCode           string    `json:"barcode"`
	BoxBarcode        string    `json:"box_barcode"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	AvailableForOrder int       `json:"availableForOrder"`
	ImagePath         string    `json:"imagePath"`
	CreatedAt         time.Time `json:"createdAt,omitempty"`
	NameJpn           string    `json:"name_jpn"`
	NameChn           string    `json:"name_chn"`
	NameKor           string    `json:"name_kor"`
	NameEng           string    `json:"name_eng"`
	Stock             []Stock   `json:"stock"`
	Tag               []Tag     `json:"tag"`
	Ingredients       string    `json:"ingredients"`
	IsBeefContained   bool      `json:"isBeefContained"`
	IsPorkContained   bool      `json:"isPorkContained"`
	IsHalal           bool      `json:"isHalal"`
	IsPlantBased      bool      `json:"isPlantBased"`
	Reasoning         string    `json:"reasoning"`
}
type Stock struct {
	StockId           string    `json:"stockId"`
	ItemId            string    `json:"itemId"`
	BoxNumber         int       `json:"boxNumber"`
	SingleNumber      int       `json:"singleNumber"`
	BundleNumber      int       `json:"bundleNumber"`
	ExpiryDate        time.Time `json:"expiryDate"`
	Location          string    `json:"location"`
	RegisteringPerson string    `json:"registeringPerson"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"createdAt,omitempty"`
}

type StockTransaction struct {
	ID              string    `json:"id"`
	ItemID          string    `json:"itemId"`
	Quantity        int       `json:"quantity"`
	TransactionType string    `json:"transactionType"` // "in" or "out"
	UserEmail       string    `json:"userEmail"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt,omitempty"`
}

// GetItemByBarcode retrieves an item by its barcode
func GetItemByBarcode(barcode string) (Item, error) {
	fmt.Println("---GETITEMBYBARCODE---", barcode)
	db := GetDBInstance(GetDBConfig())
	var item Item

	query := `SELECT item_id, code, barcode, box_barcode, name, type, available_for_order, image_path, created_at 
				FROM items 
				WHERE barcode = ? OR box_barcode = ?;`
	err := db.QueryRow(query, barcode, barcode).Scan(
		&item.ID,
		&item.Code,
		&item.BarCode,
		&item.BoxBarcode,
		&item.Name,
		&item.Type,
		&item.AvailableForOrder,
		&item.ImagePath,
		&item.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return item, errors.New("item not found")
		}
		return item, err
	}
	stocks := []Stock{}
	query = "SELECT box_number, single_number, bundle_number, stock_id, fkproduct_id,expiry_date, location, registering_person, notes, created_at FROM stocks WHERE fkproduct_id = ?"
	rows, err := db.Query(query, item.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return item, errors.New("stock not found")
		}
		return item, err
	}
	defer rows.Close()
	for rows.Next() {
		var stock Stock
		err = rows.Scan(&stock.BoxNumber, &stock.SingleNumber, &stock.BundleNumber, &stock.StockId, &stock.ItemId, &stock.ExpiryDate, &stock.Location, &stock.RegisteringPerson, &stock.Notes, &stock.CreatedAt)
		if err != nil {
			return item, err
		}
		stocks = append(stocks, stock)
	}
	if err = rows.Err(); err != nil {
		return item, err
	}
	item.Stock = stocks

	return item, nil
}

// CreateItem creates a new item in the database with tags
func CreateItem(item Item) (Item, error) {
	fmt.Println("---CREATEITEM---", item)
	db := GetDBInstance(GetDBConfig())

	// Generate a unique ID if not provided
	if item.ID == "" {
		item.ID = fmt.Sprintf("item_%d", time.Now().UnixNano())
	}

	now := time.Now()
	item.CreatedAt = now

	// Set created_at if not already set
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}

	// Start a transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return Item{}, fmt.Errorf("failed to start transaction: %v", err)
	}

	// Defer rollback in case of error
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// Insert the item
	query := "INSERT INTO items ( code, barcode, box_barcode, name, name_jpn, name_chn, name_kor, name_eng, type, available_for_order, image_path, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(query,
		item.Code,
		item.BarCode,
		item.BoxBarcode,
		item.Name,
		item.NameJpn,
		item.NameChn,
		item.NameKor,
		item.NameEng,
		item.Type,
		item.AvailableForOrder,
		item.ImagePath,
		item.CreatedAt,
	)

	if err != nil {
		return Item{}, fmt.Errorf("failed to insert item: %v", err)
	}

	// If tags are provided, associate them with the item
	if len(item.Tag) > 0 {
		// Prepare the statement for tag association
		tagStmt, err := tx.Prepare("INSERT IGNORE INTO item_tags (item_id, tag_id, created_at) VALUES (?, ?, ?)")
		if err != nil {
			return Item{}, fmt.Errorf("failed to prepare tag statement: %v", err)
		}
		defer tagStmt.Close()

		// Associate each tag with the item
		for _, tag := range item.Tag {
			_, err := tagStmt.Exec(item.ID, tag.ID, now)
			if err != nil {
				return Item{}, fmt.Errorf("failed to associate tag %s: %v", tag.ID, err)
			}
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return Item{}, fmt.Errorf("failed to commit transaction: %v", err)
	}

	// Set tx to nil to prevent rollback in defer
	tx = nil

	fmt.Printf("Successfully created item %s with %d tags\n", item.ID, len(item.Tag))
	return item, nil
}

// UpdateItem updates an existing item in the database
func UpdateItem(item Item) (Item, error) {
	fmt.Println("---UPDATEITEM---", item)

	if item.ID == "" && (item.BarCode == "" || item.Code == "") {
		return Item{}, fmt.Errorf("item ID or barcode and code are required for update")
	}

	db := GetDBInstance(GetDBConfig())

	// First check if the item exists
	var existingItem Item
	var err error

	if item.ID != "" {
		// If ID is provided, get item by ID
		existingItem, err = GetItemById(item.ID)
	} else if item.Code != "" {
		// Try to get by code
		existingItem, err = GetItemByCode(item.Code)
	}

	if err != nil {
		return Item{}, fmt.Errorf("item not found: %v", err)
	}

	// If fields are not provided in the update, keep the existing values
	if item.Name == "" {
		item.Name = existingItem.Name
	}

	if item.Type == "" {
		item.Type = existingItem.Type
	}

	if item.ImagePath == "" {
		item.ImagePath = existingItem.ImagePath
	}
	if item.BoxBarcode == "" {
		item.BoxBarcode = existingItem.BoxBarcode
	}

	// Use the existing ID to ensure we're updating the right item
	if item.ID == "" {
		item.ID = existingItem.ID
	}

	// We won't update the CreatedAt timestamp
	item.CreatedAt = existingItem.CreatedAt

	// Prepare update query
	query := `
	UPDATE items 
	SET name = ?, type = ?, name_jpn = ?, name_chn = ?, name_kor = ?, name_eng = ?, barcode = ?, box_barcode = ?,available_for_order = ?, image_path = ?
	WHERE item_id = ?`

	stmt, err := db.Prepare(query)
	if err != nil {
		return Item{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		item.Name,
		item.Type,
		item.NameJpn,
		item.NameChn,
		item.NameKor,
		item.NameEng,
		item.BarCode,
		item.BoxBarcode,
		item.AvailableForOrder,
		item.ImagePath,
		item.ID,
	)

	if err != nil {
		return Item{}, err
	}

	// Get the updated item to return
	updatedItem, err := GetItemById(item.ID)
	if err != nil {
		return item, nil // Return the input item if we can't fetch the updated one
	}

	return updatedItem, nil
}

// StockIn adds quantity to an item's stock
func StockIn(itemID string, quantity int, userID string, notes string) error {
	fmt.Println("---STOCKIN---", itemID, quantity, userID, notes)
	db := GetDBInstance(GetDBConfig())

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 1. Update item quantity
	updateQuery := "UPDATE items SET quantity_in_stock = quantity_in_stock + ?, last_updated = ? WHERE id = ?"
	_, err = tx.Exec(updateQuery, quantity, time.Now(), itemID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Create stock transaction record
	transactionID := fmt.Sprintf("transaction_%d", time.Now().UnixNano())
	transactionQuery := "INSERT INTO stock_transactions (id, item_id, quantity, type, user_id, notes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(transactionQuery, transactionID, itemID, quantity, "in", userID, notes, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// StockOut removes quantity from an item's stock
func StockOut(itemID string, quantity int, userID string, notes string) error {
	fmt.Println("---STOCKOUT---", itemID, quantity, userID, notes)
	db := GetDBInstance(GetDBConfig())

	// First check if there's enough stock
	var currentQuantity int
	queryCheck := "SELECT quantity_in_stock FROM items WHERE item_id = ?"
	err := db.QueryRow(queryCheck, itemID).Scan(&currentQuantity)
	if err != nil {
		return err
	}

	if currentQuantity < quantity {
		return errors.New("insufficient stock")
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	// 1. Update item quantity
	updateQuery := "UPDATE items SET quantity_in_stock = quantity_in_stock - ?, last_updated = ? WHERE id = ?"
	_, err = tx.Exec(updateQuery, quantity, time.Now(), itemID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 2. Create stock transaction record
	transactionQuery := "INSERT INTO stock_transactions (item_id, quantity, type, user_id, notes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(transactionQuery, itemID, quantity, "out", userID, notes, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetAllItems retrieves all items from the database with their tags
func GetAllItems() ([]Item, error) {
	fmt.Println("---GETALLITEMS---")
	db := GetDBInstance(GetDBConfig())
	var items []Item
	var itemMap = make(map[string]*Item) // Map to store items by ID for easy access

	// First query to get all items
	query := "SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(box_barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.BarCode,
			&item.BoxBarcode,
			&item.Name,
			&item.Type,
			&item.AvailableForOrder,
			&item.ImagePath,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Store a reference to the item in the map
		itemMap[item.ID] = &item
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Now fetch tags for all items in a single query
	tagQuery := `
	SELECT it.item_id, t.id, t.name, IFNULL(t.category, ''), t.created_at
	FROM item_tags it
	JOIN tags t ON it.tag_id = t.id
	`
	tagRows, err := db.Query(tagQuery)
	if err != nil {
		// Continue even if there's an error fetching tags
		fmt.Printf("Error fetching tags: %v\n", err)
	} else {
		defer tagRows.Close()

		for tagRows.Next() {
			var itemID string
			var tag Tag

			err := tagRows.Scan(
				&itemID,
				&tag.ID,
				&tag.TagName,
				// &tag.Category,
				// &tag.CreatedAt,
			)
			if err != nil {
				fmt.Printf("Error scanning tag: %v\n", err)
				continue
			}

			// Add tag to the appropriate item
			if item, exists := itemMap[itemID]; exists {
				item.Tag = append(item.Tag, tag)
			}
		}

		if err = tagRows.Err(); err != nil {
			fmt.Printf("Error in tag rows: %v\n", err)
		}
	}

	// Fetch stock information for each item
	for i, item := range items {
		stocks, err := GetStocksByItemId(item.ID)
		if err != nil {
			// Continue with empty stock if there's an error
			items[i].Stock = []Stock{}
		} else {
			items[i].Stock = stocks
		}
	}

	// Convert the map values back to a slice
	var result []Item
	for _, item := range itemMap {
		result = append(result, *item)
	}

	return result, nil
}

/**

WEIRD PROBLEM.
I JUST NEED TO GET SOME ITMES HAVING SPECIFIC TAGS
BUT TO GET THE ITEMS HAVING THOSE TAGS, I HAVE TO JOIN THE TABLES WITH THE FOREIGN KEYS.
BUT AND THEN, I DON"T KNOW IF IT's LANGAUGES's LIMIT,
I HAVE TO FETCH ALL THE TAGS WITH THE KEY.
THAT MEANS I HAVE TO DO A SAME JOB TWICE.

FIX LATER

*/
// GetItemsPaginated retrieves items from the database with pagination and tag filtering

// STUDY THIS CODE
func GetItemsPaginated(offset, limit int, tagParams []string) ([]Item, int, error) {
	defer utils.Trace()()
	utils.Info("Getting paginated items with offset: %d, limit: %d, tags: %v", offset, limit, tagParams)

	db := GetDBInstance(GetDBConfig())
	if db == nil {
		utils.Error("Failed to get database instance")
		return nil, 0, fmt.Errorf("database connection error")
	}

	// Default to a specific tag if none provided
	if len(tagParams) == 0 {
		tagParams = []string{"sp6"}
		utils.Info("No tags provided, defaulting to 'sp6'")
	}

	// Create a map to store unique items by ID
	itemMap := make(map[string]*Item)

	// Create placeholders for tag names in the IN clause
	var placeholders []string
	for range tagParams {
		placeholders = append(placeholders, "?")
	}
	placeholderStr := strings.Join(placeholders, ", ")

	// Build args for the query
	args := make([]interface{}, 0, len(tagParams)+2)
	for _, tag := range tagParams {
		args = append(args, tag)
	}

	// First, get the correct total count (items with matching tags)
	var totalCount int
	countQuery := `
		SELECT COUNT(DISTINCT i.item_id)
		FROM items i
		JOIN item_tags it ON i.item_id = it.item_id
		JOIN tags t ON it.tag_id = t.id
		WHERE t.name IN (` + placeholderStr + `)
	`

	utils.Debug("Executing count query: %s with args: %v", countQuery, args)
	err := db.QueryRow(countQuery, args...).Scan(&totalCount)
	if err != nil {
		utils.Error("Error getting total count: %v", err)
		return nil, 0, err
	}
	utils.Info("Total matching items: %d", totalCount)

	// Now add limit and offset to args
	args = append(args, limit, offset)

	// Query to get paginated items with tag filtering
	query := `
		SELECT DISTINCT i.item_id, 
		IFNULL(i.code, ''), 
		IFNULL(i.barcode, ''),
		IFNULL(i.box_barcode, ''),
		IFNULL(i.name, ''), 
		IFNULL(i.type, ''), 
		IFNULL(i.available_for_order, 0), 
		IFNULL(i.image_path, ''), 
		i.created_at,
		IFNULL(i.name_jpn, ''), 
		IFNULL(i.name_chn, ''), 
		IFNULL(i.name_kor, ''), 
		IFNULL(i.name_eng, '')
		FROM items i
		JOIN item_tags it ON i.item_id = it.item_id
		JOIN tags t ON it.tag_id = t.id
		WHERE t.name IN (` + placeholderStr + `)
		GROUP BY i.item_id
		ORDER BY i.created_at DESC
		LIMIT ? OFFSET ?
	`

	utils.Debug("Executing query: %s with args: %v", query, args)
	rows, err := db.Query(query, args...)
	if err != nil {
		utils.Error("Error executing items query: %v", err)
		return nil, 0, err
	}
	defer rows.Close()

	var items []Item
	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.BarCode,
			&item.BoxBarcode,
			&item.Name,
			&item.Type,
			&item.AvailableForOrder,
			&item.ImagePath,
			&item.CreatedAt,
			&item.NameJpn,
			&item.NameChn,
			&item.NameKor,
			&item.NameEng,
		)
		if err != nil {
			utils.Error("Error scanning item row: %v", err)
			return nil, 0, err
		}

		utils.Debug("Found item: ID=%s, Name=%s", item.ID, item.Name)
		itemMap[item.ID] = &item
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		utils.Error("Error in rows iteration: %v", err)
		return nil, 0, err
	}
	fmt.Println("---ITEMS---", items)

	// If we have items, fetch their tags and stocks
	if len(items) > 0 {
		// Get all item IDs for batch fetching
		var itemIDs []string
		for _, item := range items {
			itemIDs = append(itemIDs, item.ID)
		}

		// Create placeholders for item IDs
		itemPlaceholders := make([]string, len(itemIDs))
		itemArgs := make([]interface{}, len(itemIDs))
		for i, id := range itemIDs {
			itemPlaceholders[i] = "?"
			itemArgs[i] = id
		}
		itemPlaceholderStr := strings.Join(itemPlaceholders, ", ")

		// Fetch tags for all items in one query
		tagQuery := `
			SELECT it.item_id, t.id, t.name
			FROM item_tags it
			JOIN tags t ON it.tag_id = t.id
			WHERE it.item_id IN (` + itemPlaceholderStr + `)
		`

		utils.Debug("Executing tag query: %s with args: %v", tagQuery, itemArgs)
		tagRows, err := db.Query(tagQuery, itemArgs...)
		if err != nil {
			utils.Warn("Error fetching tags: %v", err)
			// Continue without tags if there's an error
		} else {
			defer tagRows.Close()
			for tagRows.Next() {
				var itemID string
				var tag Tag
				err := tagRows.Scan(&itemID, &tag.ID, &tag.TagName)
				if err != nil {
					utils.Warn("Error scanning tag: %v", err)
					continue
				}

				if item, exists := itemMap[itemID]; exists {
					item.Tag = append(item.Tag, tag)
				}
			}

			if err = tagRows.Err(); err != nil {
				utils.Warn("Error in tag rows iteration: %v", err)
			}
		}

		// Fetch stock for each item
		for i, item := range items {
			utils.Debug("Fetching stocks for item: %s", item.ID)
			stocks, err := GetStocksByItemId(item.ID)
			if err != nil {
				utils.Warn("Error fetching stocks for item %s: %v", item.ID, err)
				items[i].Stock = []Stock{} // Empty stock array if error
			} else {
				items[i].Stock = stocks
				utils.Debug("Found %d stocks for item %s", len(stocks), item.ID)
			}
		}
	}

	// Rebuild items list from the map to ensure all data is included
	var result []Item
	for _, item := range itemMap {
		result = append(result, *item)
	}

	utils.Info("Successfully retrieved %d items", len(result))
	return result, totalCount, nil
}

func GetItemByCode(code string) (Item, error) {
	fmt.Println("---GETITEMBYCODE---", code)
	db := GetDBInstance(GetDBConfig())
	var item Item
	query := "SELECT * FROM items WHERE code = ?"
	err := db.QueryRow(query, code).Scan(&item.ID, &item.Code, &item.BarCode, &item.BoxBarcode, &item.Name,
		&item.Type, &item.AvailableForOrder, &item.ImagePath, &item.CreatedAt)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func AddStock(stock Stock) error {
	fmt.Println("---ADDSTOCK---", stock)
	db := GetDBInstance(GetDBConfig())
	query := "INSERT INTO stocks ( fkproduct_id, box_number, single_number, bundle_number, expiry_date, location, registering_person, notes) VALUES (?, ?, ?, ?, ?, ?, ?, ? )"
	_, err := db.Exec(query, stock.ItemId, stock.BoxNumber, stock.SingleNumber, stock.BundleNumber, stock.ExpiryDate, stock.Location, stock.RegisteringPerson, stock.Notes)
	if err != nil {
		return err
	}
	return nil
}
func SaveStockTransaction(transaction StockTransaction) error {
	fmt.Println("---SAVESTOCKTRANSACTION---", transaction)
	db := GetDBInstance(GetDBConfig())
	fmt.Println("@@@transaction.UserEmail", transaction.UserEmail)
	query := "INSERT INTO stock_transactions ( fkitem_id, quantity, transaction_type, fkuser_email) VALUES (?, ?, ?, ?)"
	_, err := db.Exec(query, transaction.ItemID, transaction.Quantity, transaction.TransactionType, transaction.UserEmail)
	if err != nil {
		return err
	}
	return nil
}

func GetItemById(id string) (Item, error) {
	defer utils.Trace()()
	utils.Info("Getting item by ID: %s", id)

	if id == "" {
		utils.Error("Empty item ID provided to GetItemById")
		return Item{}, fmt.Errorf("empty item ID")
	}

	db := GetDBInstance(GetDBConfig())
	if db == nil {
		utils.Error("Failed to get database instance")
		return Item{}, fmt.Errorf("database connection error")
	}

	var item Item
	query := "SELECT item_id, code, IFNULL(barcode, ''), IFNULL(box_barcode, ''), IFNULL(name, ''), IFNULL(type, ''), " +
		"IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at, " +
		"IFNULL(name_jpn, ''), IFNULL(name_chn, ''), IFNULL(name_kor, ''), IFNULL(name_eng, '') " +
		"FROM items WHERE item_id = ?"
	fmt.Println("---QUERY---", query)
	utils.Debug("Executing query: %s with item ID: %s", query, id)

	err := db.QueryRow(query, id).Scan(
		&item.ID,
		&item.Code,
		&item.BarCode,
		&item.BoxBarcode,
		&item.Name,
		&item.Type,
		&item.AvailableForOrder,
		&item.ImagePath,
		&item.CreatedAt,
		&item.NameJpn,
		&item.NameChn,
		&item.NameKor,
		&item.NameEng,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			utils.Warn("No item found with ID: %s", id)
			return Item{}, fmt.Errorf("item not found")
		}
		utils.Error("Error querying item with ID %s: %v", id, err)
		return Item{}, err
	}

	utils.Info("Successfully retrieved item: ID=%s, Name='%s', Code=%s",
		item.ID, item.Name, item.Code)
	return item, nil
}

func GetStocksByItemId(itemId string) ([]Stock, error) {
	defer utils.Trace()()
	utils.Info("Getting stocks for item ID: %s", itemId)

	if itemId == "" {
		utils.Error("Empty item ID provided to GetStocksByItemId")
		return nil, fmt.Errorf("empty item ID")
	}

	db := GetDBInstance(GetDBConfig())
	if db == nil {
		utils.Error("Failed to get database instance")
		return nil, fmt.Errorf("database connection error")
	}

	var stocks []Stock
	query := "SELECT stock_id, fkproduct_id, box_number, single_number, bundle_number, expiry_date, location, registering_person, notes, created_at FROM stocks WHERE fkproduct_id = ?"
	utils.Debug("Executing query: %s with item ID: %s", query, itemId)

	rows, err := db.Query(query, itemId)
	if err != nil {
		utils.Error("Error querying stocks for item %s: %v", itemId, err)
		return nil, err
	}
	defer rows.Close()

	stockCount := 0
	for rows.Next() {
		var stock Stock
		err := rows.Scan(
			&stock.StockId,
			&stock.ItemId,
			&stock.BoxNumber,
			&stock.SingleNumber,
			&stock.BundleNumber,
			&stock.ExpiryDate,
			&stock.Location,
			&stock.RegisteringPerson,
			&stock.Notes,
			&stock.CreatedAt,
		)
		if err != nil {
			utils.Error("Error scanning stock row: %v", err)
			return nil, err
		}
		utils.Debug("Found stock: ID=%s, BoxNumber=%d, Location=%s",
			stock.StockId, stock.BoxNumber, stock.Location)
		stocks = append(stocks, stock)
		stockCount++
	}

	if err = rows.Err(); err != nil {
		utils.Error("Error iterating through rows: %v", err)
		return nil, err
	}

	utils.Info("Retrieved %d stocks for item %s", stockCount, itemId)
	return stocks, nil
}
func UpdateStock(stockId string, stockType string, quantity int) error {
	db := GetDBInstance(GetDBConfig())
	query := "UPDATE stocks SET box_number = box_number - ? WHERE stock_id = ?"
	_, err := db.Exec(query, quantity, stockId)
	if err != nil {
		return err
	}
	return nil
}

func RemoveStock(stockId string) error {
	db := GetDBInstance(GetDBConfig())
	// delete row from stocks table
	query := "DELETE FROM stocks WHERE stock_id = ?"
	_, err := db.Exec(query, stockId)
	if err != nil {
		return err
	}
	return nil
}

// SearchItemsByField searches for items using LIKE query on the specified field
func SearchItemsByField(searchType string, value string) ([]Item, error) {
	fmt.Println("---SEARCHITEMSBYFIELD---", searchType, value)
	db := GetDBInstance(GetDBConfig())
	var items []Item
	var itemMap = make(map[string]*Item) // Map to store items by ID for easy access
	var query string

	// Add % wildcards for LIKE query
	searchValue := "%" + value + "%"

	// Determine which field to search
	switch searchType {
	case "code":
		query = "SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items WHERE code LIKE ?"
	case "barcode":
		query = "SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items WHERE barcode LIKE ?"
	case "name":
		query = "SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items WHERE name LIKE ?"
	default:
		return nil, fmt.Errorf("invalid search type: %s", searchType)
	}

	rows, err := db.Query(query, searchValue)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item Item
		err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.BarCode,
			&item.BoxBarcode,
			&item.Name,
			&item.Type,
			&item.AvailableForOrder,
			&item.ImagePath,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Store a reference to the item in the map
		itemMap[item.ID] = &item
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	// Now fetch tags for all items in a single query
	if len(items) > 0 {
		// Build a list of item IDs for the IN clause
		var itemIDs []string
		for _, item := range items {
			itemIDs = append(itemIDs, item.ID)
		}

		// Create placeholders for SQL query
		placeholders := ""
		args := make([]interface{}, len(itemIDs))

		for i, itemID := range itemIDs {
			if i > 0 {
				placeholders += ", "
			}
			placeholders += "?"
			args[i] = itemID
		}

		tagQuery := `
		SELECT it.item_id, t.id, t.name, IFNULL(t.category, ''), t.created_at
		FROM item_tags it
		JOIN tags t ON it.tag_id = t.id
		WHERE it.item_id IN (` + placeholders + `)`

		tagRows, err := db.Query(tagQuery, args...)
		if err != nil {
			// Continue even if there's an error fetching tags
			fmt.Printf("Error fetching tags: %v\n", err)
		} else {
			defer tagRows.Close()

			for tagRows.Next() {
				var itemID string
				var tag Tag

				err := tagRows.Scan(
					&itemID,
					&tag.ID,
					&tag.TagName,
					// &tag.Category,
					// &tag.CreatedAt,
				)
				if err != nil {
					fmt.Printf("Error scanning tag: %v\n", err)
					continue
				}

				// Add tag to the appropriate item
				if item, exists := itemMap[itemID]; exists {
					item.Tag = append(item.Tag, tag)
				}
			}

			if err = tagRows.Err(); err != nil {
				fmt.Printf("Error in tag rows: %v\n", err)
			}
		}

		// Fetch stock information for each item
		for i, item := range items {
			stocks, err := GetStocksByItemId(item.ID)
			if err != nil {
				// Continue with empty stock if there's an error
				items[i].Stock = []Stock{}
			} else {
				items[i].Stock = stocks
			}
		}
	}

	// Convert the map values back to a slice with updated references
	var result []Item
	for _, item := range itemMap {
		result = append(result, *item)
	}

	return result, nil
}

// GetTagsByItemId retrieves all tags for a given item ID
func GetTagsByItemId(itemId string) ([]Tag, error) {
	db := GetDBInstance(GetDBConfig())
	var tags []Tag

	query := `
		SELECT t.id, t.name 
		FROM tags t
		JOIN item_tags it ON t.id = it.tag_id
		WHERE it.item_id = ?
	`
	rows, err := db.Query(query, itemId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.TagName); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

// ItemWithDaysToExpiry represents an item with days until expiry calculation
type ItemWithDaysToExpiry struct {
	Item         Item   `json:"item"`
	DaysToExpiry int    `json:"daysToExpiry"`
	StockId      string `json:"stockId"`
}

// GetItemsExpiringWithinDays retrieves items that are expiring within the specified number of days
func GetItemsExpiringWithinDays(withinDays int) ([]ItemWithDaysToExpiry, error) {
	defer utils.Trace()()
	utils.Info("Getting items expiring within %d days", withinDays)

	db := GetDBInstance(GetDBConfig())
	if db == nil {
		utils.Error("Failed to get database instance")
		return nil, fmt.Errorf("database connection error")
	}

	// Query to get items with their stock expiry dates and calculate days to expiry
	query := `
		SELECT 
			i.item_id, 
			IFNULL(i.code, ''), 
			IFNULL(i.barcode, ''),
			IFNULL(i.box_barcode, ''),
			IFNULL(i.name, ''), 
			IFNULL(i.type, ''), 
			IFNULL(i.available_for_order, 0), 
			IFNULL(i.image_path, ''), 
			i.created_at,
			IFNULL(i.name_jpn, ''), 
			IFNULL(i.name_chn, ''), 
			IFNULL(i.name_kor, ''), 
			IFNULL(i.name_eng, ''),
			s.stock_id,
			s.expiry_date,
			DATEDIFF(s.expiry_date, CURDATE()) as days_to_expiry
		FROM items i
		JOIN stocks s ON i.item_id = s.fkproduct_id
		WHERE DATEDIFF(s.expiry_date, CURDATE()) <= ? 
		AND DATEDIFF(s.expiry_date, CURDATE()) >= 0
		ORDER BY days_to_expiry ASC
	`

	utils.Debug("Executing query: %s with withinDays: %d", query, withinDays)
	rows, err := db.Query(query, withinDays)
	if err != nil {
		utils.Error("Error executing expiry query: %v", err)
		return nil, err
	}
	defer rows.Close()

	var results []ItemWithDaysToExpiry
	itemMap := make(map[string]*Item) // To avoid duplicate items

	for rows.Next() {
		var item Item
		var stockId string
		var expiryDate time.Time
		var daysToExpiry int

		err := rows.Scan(
			&item.ID,
			&item.Code,
			&item.BarCode,
			&item.BoxBarcode,
			&item.Name,
			&item.Type,
			&item.AvailableForOrder,
			&item.ImagePath,
			&item.CreatedAt,
			&item.NameJpn,
			&item.NameChn,
			&item.NameKor,
			&item.NameEng,
			&stockId,
			&expiryDate,
			&daysToExpiry,
		)
		if err != nil {
			utils.Error("Error scanning expiry row: %v", err)
			return nil, err
		}

		// Create the result item
		result := ItemWithDaysToExpiry{
			Item:         item,
			DaysToExpiry: daysToExpiry,
			StockId:      stockId,
		}

		// If we haven't seen this item before, get its tags and stocks
		if _, exists := itemMap[item.ID]; !exists {
			// Get tags for this item
			tags, err := GetTagsByItemId(item.ID)
			if err != nil {
				utils.Warn("Error fetching tags for item %s: %v", item.ID, err)
			} else {
				result.Item.Tag = tags
			}

			// Get all stocks for this item
			stocks, err := GetStocksByItemId(item.ID)
			if err != nil {
				utils.Warn("Error fetching stocks for item %s: %v", item.ID, err)
			} else {
				result.Item.Stock = stocks
			}

			itemMap[item.ID] = &result.Item
		} else {
			// Use cached item data
			result.Item = *itemMap[item.ID]
		}

		results = append(results, result)
		utils.Debug("Found expiring item: ID=%s, Name=%s, DaysToExpiry=%d",
			item.ID, item.Name, daysToExpiry)
	}

	if err = rows.Err(); err != nil {
		utils.Error("Error in rows iteration: %v", err)
		return nil, err
	}

	utils.Info("Successfully retrieved %d expiring items", len(results))
	return results, nil
}
