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
	BarCode           string    `json:"barCode"`
	Name              string    `json:"name"`
	Type              string    `json:"type"`
	AvailableForOrder int       `json:"availableForOrder"`
	ImagePath         string    `json:"imagePath"`
	CreatedAt         time.Time `json:"createdAt,omitempty"`
	Stock             []Stock   `json:"stock"`
}
type Stock struct {
	StockId        string    `json:"stockId"`
	ItemId         string    `json:"itemId"`
	BoxNumber      int       `json:"boxNumber"`
	SingleNumber   int       `json:"singleNumber"`
	BundleNumber   int       `json:"bundleNumber"`
	RegisteredDate time.Time `json:"registeredDate"`
	Notes          string    `json:"notes"`
	CreatedAt      time.Time `json:"createdAt,omitempty"`
}

type StockTransaction struct {
	ID              string    `json:"id"`
	ItemID          string    `json:"itemId"`
	Quantity        int       `json:"quantity"`
	TransactionType string    `json:"transactionType"` // "in" or "out"
	UserID          string    `json:"userId"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"createdAt"`
}

// GetItemByBarcode retrieves an item by its barcode
func GetItemByBarcode(barcode string) (Item, error) {
	fmt.Println("---GETITEMBYBARCODE---", barcode)
	db := GetDBInstance(GetDBConfig())
	var item Item

	query := "SELECT id, code, barcode, name, type, available_for_order, image_path, created_at FROM items WHERE barcode = ?"
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

	return item, nil
}

// CreateItem creates a new item in the database
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
	queryCheck := "SELECT quantity_in_stock FROM items WHERE id = ?"
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
	transactionID := fmt.Sprintf("transaction_%d", time.Now().UnixNano())
	transactionQuery := "INSERT INTO stock_transactions (id, item_id, quantity, type, user_id, notes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err = tx.Exec(transactionQuery, transactionID, itemID, quantity, "out", userID, notes, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// GetAllItems retrieves all items from the database
func GetAllItems() ([]Item, error) {
	fmt.Println("---GETALLITEMS---")
	db := GetDBInstance(GetDBConfig())
	var items []Item

	query := "SELECT id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items"
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
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
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
	query := "INSERT INTO stocks (stock_id, item_id, box_number, single_number, bundle_number, registered_date, notes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)"
	_, err := db.Exec(query, stock.StockId, stock.ItemId, stock.BoxNumber, stock.SingleNumber, stock.BundleNumber, stock.RegisteredDate, stock.Notes, stock.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}
func SaveStockTransaction(transaction StockTransaction) error {
	fmt.Println("---SAVESTOCKTRANSACTION---", transaction)
	db := GetDBInstance(GetDBConfig())
	query := "INSERT INTO stock_transactions (id, item_id, quantity, type, user_id, notes, created_at) VALUES (?, ?, ?, ?, ?, ?, ?)"
	_, err := db.Exec(query, transaction.ID, transaction.ItemID, transaction.Quantity, transaction.TransactionType, transaction.UserID, transaction.Notes, transaction.CreatedAt)
	if err != nil {
		return err
	}
	return nil
}

func GetItemById(id string) (Item, error) {
	fmt.Println("---GETITEMBYID---", id)
	db := GetDBInstance(GetDBConfig())
	var item Item
	query := "SELECT * FROM items WHERE id = ?"
	err := db.QueryRow(query, id).Scan(&item.ID, &item.Code, &item.BarCode, &item.Name, &item.Type, &item.AvailableForOrder, &item.ImagePath, &item.CreatedAt)
	if err != nil {
		return Item{}, err
	}
	return item, nil
}

func GetStocksByItemId(stockId string) ([]Stock, error) {
	fmt.Println("---GETSTOCKBYITEMID---", stockId)
	db := GetDBInstance(GetDBConfig())
	var stocks []Stock
	query := "SELECT * FROM stocks WHERE stock_id = ?"
	rows, err := db.Query(query, stockId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var stock Stock
		err := rows.Scan(&stock.StockId, &stock.ItemId, &stock.BoxNumber, &stock.SingleNumber, &stock.BundleNumber, &stock.RegisteredDate, &stock.Notes, &stock.CreatedAt)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, stock)
	}
	return stocks, nil
}

func RemoveStock(stockId string, stockType string, quantity int) error {
	db := GetDBInstance(GetDBConfig())
	query := "UPDATE stocks SET box_number = box_number - ?, single_number = single_number - ?, bundle_number = bundle_number - ? WHERE stock_id = ?"
	_, err := db.Exec(query, quantity, quantity, quantity, stockId)
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
	var query string

	// Add % wildcards for LIKE query
	searchValue := "%" + value + "%"

	// Determine which field to search
	switch searchType {
	case "code":
		query = "SELECT id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items WHERE code LIKE ?"
	case "barcode":
		query = "SELECT id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items WHERE barcode LIKE ?"
	case "name":
		query = "SELECT id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(name, ''), IFNULL(type, ''), IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at FROM items WHERE name LIKE ?"
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
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return items, nil
}
