package models

import (
	"fmt"
	"time"
)

// Barcode represents a barcode entity in the system
type Barcode struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

// SaveBarcode stores a new barcode in the database
func SaveBarcode(code string, userEmail string) (Barcode, error) {

	if code == "" {
		return Barcode{}, fmt.Errorf("barcode cannot be empty")
	}

	// Get database instance
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		return Barcode{}, fmt.Errorf("database connection error")
	}

	// Create barcode entity
	barcode := Barcode{
		ID:        fmt.Sprintf("barcode_%d", time.Now().UnixNano()),
		Code:      code,
		CreatedAt: time.Now(),
	}

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return Barcode{}, err
	}

	// Defer a rollback in case anything fails
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	// Check if the barcode already exists
	var exists int
	existsQuery := "SELECT COUNT(*) FROM barcodes WHERE barcode = ?"
	err = tx.QueryRow(existsQuery, code).Scan(&exists)
	if err != nil {
		return Barcode{}, fmt.Errorf("error checking existing barcode: %v", err)
	}

	if exists > 0 {
		return Barcode{}, fmt.Errorf("barcode %s already exists", code)
	}

	// Insert the barcode
	insertQuery := "INSERT INTO barcodes (barcode, created_at) VALUES (?, ?)"
	_, err = tx.Exec(insertQuery, barcode.Code, barcode.CreatedAt)
	if err != nil {
		return Barcode{}, fmt.Errorf("error saving barcode: %v", err)
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		return Barcode{}, fmt.Errorf("error committing barcode transaction: %v", err)
	}

	// Set tx to nil to prevent the deferred rollback
	tx = nil

	return barcode, nil
}

// GetBarcodeByCode retrieves a barcode by its code
func GetBarcodeByCode(code string) (Barcode, error) {

	if code == "" {
		return Barcode{}, fmt.Errorf("barcode code cannot be empty")
	}

	// Get database instance
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		return Barcode{}, fmt.Errorf("database connection error")
	}

	var barcode Barcode
	query := "SELECT id, code, created_at FROM barcodes WHERE code = ?"
	err := db.QueryRow(query, code).Scan(
		&barcode.ID,
		&barcode.Code,
		&barcode.CreatedAt,
	)
	if err != nil {
		return Barcode{}, fmt.Errorf("error retrieving barcode: %v", err)
	}

	return barcode, nil
}
