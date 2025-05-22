package models

import (
	"fmt"
	"time"

	"github.com/jimyeongjung/owlverload_api/utils"
)

// Barcode represents a barcode entity in the system
type Barcode struct {
	ID        string    `json:"id"`
	Code      string    `json:"code"`
	CreatedAt time.Time `json:"createdAt"`
}

// SaveBarcode stores a new barcode in the database
func SaveBarcode(code string, userEmail string) (Barcode, error) {
	defer utils.Trace()()
	utils.Info("Saving barcode: %s", code)

	if code == "" {
		utils.Error("Empty barcode provided to SaveBarcode")
		return Barcode{}, fmt.Errorf("barcode cannot be empty")
	}

	// Get database instance
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		utils.Error("Failed to get database instance")
		return Barcode{}, fmt.Errorf("database connection error")
	}

	// Create barcode entity
	barcode := Barcode{
		ID:        fmt.Sprintf("barcode_%d", time.Now().UnixNano()),
		Code:      code,
		CreatedAt: time.Now(),
	}

	// Start a transaction
	utils.Info("Starting transaction to save barcode")
	tx, err := db.Begin()
	if err != nil {
		utils.Error("Failed to start transaction: %v", err)
		return Barcode{}, err
	}

	// Defer a rollback in case anything fails
	defer func() {
		if tx != nil {
			utils.Warn("Rolling back transaction - this happens if there was an error")
			tx.Rollback()
		}
	}()

	// Check if the barcode already exists
	var exists int
	existsQuery := "SELECT COUNT(*) FROM barcodes WHERE barcode = ?"
	err = tx.QueryRow(existsQuery, code).Scan(&exists)
	if err != nil {
		utils.Error("Error checking if barcode exists: %v", err)
		return Barcode{}, fmt.Errorf("error checking existing barcode: %v", err)
	}

	if exists > 0 {
		utils.Warn("Barcode %s already exists", code)
		return Barcode{}, fmt.Errorf("barcode %s already exists", code)
	}

	// Insert the barcode
	insertQuery := "INSERT INTO barcodes (barcode, created_at) VALUES (?, ?)"
	_, err = tx.Exec(insertQuery, barcode.Code, barcode.CreatedAt)
	if err != nil {
		utils.Error("Error inserting barcode: %v", err)
		return Barcode{}, fmt.Errorf("error saving barcode: %v", err)
	}

	// Commit the transaction
	utils.Info("Committing transaction")
	err = tx.Commit()
	if err != nil {
		utils.Error("Error committing transaction: %v", err)
		return Barcode{}, fmt.Errorf("error committing barcode transaction: %v", err)
	}

	// Set tx to nil to prevent the deferred rollback
	tx = nil

	utils.Info("Successfully saved barcode: %s with ID: %s", barcode.Code, barcode.ID)
	return barcode, nil
}

// GetBarcodeByCode retrieves a barcode by its code
func GetBarcodeByCode(code string) (Barcode, error) {
	defer utils.Trace()()
	utils.Info("Getting barcode by code: %s", code)

	if code == "" {
		utils.Error("Empty code provided to GetBarcodeByCode")
		return Barcode{}, fmt.Errorf("barcode code cannot be empty")
	}

	// Get database instance
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		utils.Error("Failed to get database instance")
		return Barcode{}, fmt.Errorf("database connection error")
	}

	var barcode Barcode
	query := "SELECT id, code, created_at FROM barcodes WHERE code = ?"

	utils.Debug("Executing query: %s with code: %s", query, code)

	err := db.QueryRow(query, code).Scan(
		&barcode.ID,
		&barcode.Code,
		&barcode.CreatedAt,
	)

	if err != nil {
		utils.Error("Error retrieving barcode: %v", err)
		return Barcode{}, fmt.Errorf("error retrieving barcode: %v", err)
	}

	utils.Info("Successfully retrieved barcode: %s", barcode.Code)
	return barcode, nil
}
