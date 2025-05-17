package models

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Item struct {
	ID                string    `json:"id"`
	Code              string    `json:"code"`
	BarCode           string    `json:"barcode"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	AvailableForOrder int       `json:"availableForOrder"`
	ImagePath         string    `json:"imagePath"`
	CreatedAt         time.Time `json:"createdAt,omitempty"`
	NameJpn           string    `json:"nameJpn"`
	NameChn           string    `json:"nameChn"`
	NameKor           string    `json:"nameKor"`
	NameEng           string    `json:"nameEng"`
	Stock             []Stock   `json:"stock"`
	Tag               []Tag     `json:"tag"`
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

	query := "SELECT item_id, code, barcode, name, type, available_for_order, image_path, created_at FROM items WHERE barcode = ?"
	err := db.QueryRow(query, barcode).Scan(
		&item.ID,
		&item.Code,
		&item.BarCode,
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

// CreateItem creates a new item in the database
func CreateItem(item Item) (Item, error) {
	fmt.Println("---CREATEITEM---", item)
	db := GetDBInstance(GetDBConfig())

	// Generate a unique ID if not provided
	if item.ID == "" {
		item.ID = fmt.Sprintf("item_%id", time.Now().UnixNano())
	}

	now := time.Now()
	item.CreatedAt = now

	// Set created_at if not already set
	if item.CreatedAt.IsZero() {
		item.CreatedAt = now
	}

	query := "INSERT INTO items (id, code, bar_code, name, type, available_for_order, image_path, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	stmt, err := db.Prepare(query)
	if err != nil {
		return Item{}, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		item.ID,
		item.Code,
		item.BarCode,
		item.Name,
		item.Type,
		item.AvailableForOrder,
		item.ImagePath,
		item.CreatedAt,
	)

	if err != nil {
		return Item{}, err
	}

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

	// Use the existing ID to ensure we're updating the right item
	if item.ID == "" {
		item.ID = existingItem.ID
	}

	// We won't update the CreatedAt timestamp
	item.CreatedAt = existingItem.CreatedAt

	// Prepare update query
	query := `
	UPDATE items 
	SET name = ?, type = ?, name_jpn = ?, name_chn = ?, name_kor = ?, name_eng = ?, barcode = ?, available_for_order = ?, image_path = ?
	WHERE id = ?`

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
	query := "SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items"
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

func GetItemByCode(code string) (Item, error) {
	fmt.Println("---GETITEMBYCODE---", code)
	db := GetDBInstance(GetDBConfig())
	var item Item
	query := "SELECT * FROM items WHERE code = ?"
	err := db.QueryRow(query, code).Scan(&item.ID, &item.Code, &item.BarCode, &item.Name,
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
	fmt.Println("---GETITEMBYID---", id)
	db := GetDBInstance(GetDBConfig())
	var item Item
	query := "SELECT item_id, code, IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at, IFNULL(name_jpn, ''), IFNULL(name_chn, ''), IFNULL(name_kor, ''), IFNULL(name_eng, '')	 FROM items WHERE item_id = ?"
	err := db.QueryRow(query, id).Scan(&item.ID, &item.Code, &item.BarCode, &item.Name, &item.Type, &item.AvailableForOrder, &item.ImagePath, &item.CreatedAt, &item.NameJpn, &item.NameChn, &item.NameKor, &item.NameEng)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func GetStocksByItemId(stockId string) ([]Stock, error) {
	fmt.Println("---GETSTOCKBYITEMID---", stockId)
	db := GetDBInstance(GetDBConfig())
	var stocks []Stock
	query := "SELECT * FROM stocks WHERE fkproduct_id = ?"
	rows, err := db.Query(query, stockId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var stock Stock
		err := rows.Scan(&stock.StockId, &stock.ItemId, &stock.BoxNumber, &stock.SingleNumber, &stock.BundleNumber, &stock.ExpiryDate, &stock.Location, &stock.RegisteringPerson, &stock.Notes, &stock.CreatedAt)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}
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
