package v1

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jimyeongjung/owlverload_api/models"
)

type StockType string

const (
	StockTypeBox    StockType = "BOX"
	StockTypeBundle StockType = "BUNDLE"
	StockTypePCS    StockType = "PCS"
)

type Stock struct {
	StockId           string    `json:"stock_id"`
	ItemId            string    `json:"item_id"`
	StockType         StockType `json:"stock_type"`
	BoxNumber         int       `json:"box_number"`
	PCSNumber         int       `json:"pcs_number"`
	BundleNumber      int       `json:"bundle_number"`
	ExpiryDate        time.Time `json:"expiry_date"`
	Location          string    `json:"location"`
	RegisteringPerson string    `json:"registering_person"`
	Notes             string    `json:"notes"`
	CreatedAt         time.Time `json:"created_at,omitempty"`
	DiscountRate      int       `json:"discount_rate"`
}
type ExpiredStock struct {
	Stock           Stock `json:"stock"`
	DaysSinceExpiry int   `json:"days_since_expiry"`
}

// DeleteStockByID deletes a stock record by its stock_id
func DeleteStockByIDWithTx(tx *sql.Tx, stockID int) error {
	if stockID == 0 {
		return fmt.Errorf("empty stock ID")
	}

	query := `DELETE FROM stocks WHERE stock_id = ?`
	result, err := tx.Exec(query, stockID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("stock not found")
	}
	return nil
}
func DeleteStockByID(stockID int) error {
	if stockID == 0 {
		return fmt.Errorf("empty stock ID")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `DELETE FROM stocks WHERE stock_id = ?`
	result, err := db.Exec(query, stockID)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("stock not found")
	}
	return nil
}

// GetStockByID retrieves a stock record by its stock_id
func GetStockByID(stockID string) (Stock, error) {
	if stockID == "" {
		return Stock{}, fmt.Errorf("empty stock ID")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	var s Stock
	query := `SELECT stock_id, fkproduct_id, stock_type, box_number, pcs_number, bundle_number, expiry_date,
		IFNULL(location, ''), IFNULL(registering_person, ''), IFNULL(notes, ''), IFNULL(discount_rate, 0), created_at
		FROM stocks WHERE stock_id = ?`
	err := db.QueryRow(query, stockID).Scan(
		&s.StockId, &s.ItemId, &s.StockType, &s.BoxNumber, &s.PCSNumber, &s.BundleNumber,
		&s.ExpiryDate, &s.Location, &s.RegisteringPerson, &s.Notes, &s.DiscountRate, &s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Stock{}, errors.New("stock not found")
		}
		return Stock{}, err
	}
	return s, nil
}

// GetStocksByProductID retrieves all stock records for a product (item)
func GetStocksByProductID(productID string) ([]Stock, error) {
	if productID == "" {
		return nil, fmt.Errorf("empty product ID")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `SELECT stock_id, fkproduct_id, stock_type, box_number, pcs_number, bundle_number, expiry_date,
		IFNULL(location, ''), IFNULL(registering_person, ''), IFNULL(notes, ''), IFNULL(discount_rate, 0), created_at
		FROM stocks WHERE fkproduct_id = ?`
	rows, err := db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []Stock
	for rows.Next() {
		var s Stock
		err := rows.Scan(
			&s.StockId, &s.ItemId, &s.StockType, &s.BoxNumber, &s.PCSNumber, &s.BundleNumber,
			&s.ExpiryDate, &s.Location, &s.RegisteringPerson, &s.Notes, &s.DiscountRate, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stocks, nil
}

// GetStocksByExpiryDateRange retrieves all stocks whose expiry_date is between startDate and endDate (inclusive)
func GetStocksByExpiryDateRange(startDate, endDate time.Time) ([]Stock, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `SELECT stock_id, fkproduct_id, stock_type, box_number, pcs_number, bundle_number, expiry_date,
		IFNULL(location, ''), IFNULL(registering_person, ''), IFNULL(notes, ''), IFNULL(discount_rate, 0), created_at
		FROM stocks WHERE expiry_date >= ? AND expiry_date <= ?
		ORDER BY expiry_date ASC`
	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stocks []Stock
	for rows.Next() {
		var s Stock
		err := rows.Scan(
			&s.StockId, &s.ItemId, &s.StockType, &s.BoxNumber, &s.PCSNumber, &s.BundleNumber,
			&s.ExpiryDate, &s.Location, &s.RegisteringPerson, &s.Notes, &s.DiscountRate, &s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return stocks, nil
}

func GetExpiredStocksByDays(days int) ([]ExpiredStock, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		SELECT stock_id,
		fkproduct_id,
		stock_type,
		box_number,
		pcs_number,
		bundle_number,
		expiry_date,
		IFNULL(location, ''),
		IFNULL(registering_person, ''),
		IFNULL(notes, ''), IFNULL(discount_rate, 0), created_at
		FROM stocks 
		WHERE expiry_date <= DATE_SUB(CURDATE(), INTERVAL ? DAY)
		ORDER BY expiry_date ASC
	`
	rows, err := db.Query(query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var expiredStocks []ExpiredStock
	for rows.Next() {
		var s ExpiredStock
		err := rows.Scan(
			&s.Stock.StockId,
			&s.Stock.ItemId,
			&s.Stock.StockType,
			&s.Stock.BoxNumber,
			&s.Stock.PCSNumber,
			&s.Stock.BundleNumber,
			&s.Stock.ExpiryDate,
			&s.Stock.Location,
			&s.Stock.RegisteringPerson,
			&s.Stock.Notes,
			&s.Stock.DiscountRate,
			&s.Stock.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		s.DaysSinceExpiry = int(time.Since(s.Stock.ExpiryDate).Hours() / 24)
		expiredStocks = append(expiredStocks, s)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return expiredStocks, nil
}
