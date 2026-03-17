package v1

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jimyeongjung/owlverload_api/models"
)

// db  scheme
// inventory_log_id int AUTO_INCREMENT PRIMARY KEY,
//
//	event_type enum('stock_in', 'stock_out', 'expired', 'damaged', 'sold', 'discounted') NOT NULL,
//	product_id INT NOT NULL,
//	product_code VARCHAR(50) NOT NULL,
//	product_name VARCHAR(255) NOT NULL,
//	product_image_path VARCHAR(255) NULL,
//	stock_id INT NOT NULL,
//	stock_type ENUM('BOX', 'PCS') NOT NULL,
//	stock_quantity INT NOT NULL,
//	expiry_date DATE,
//	original_price DECIMAL(10, 2),
//	discounted_price DECIMAL(10, 2),
//	discount_rate DECIMAL(5, 2),
//	performer_id INT NOT NULL,
//	performer_name VARCHAR(50) NOT NULL,
//	performer_email VARCHAR(100) NOT NULL,
//	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP

type EventType string

const (
	EventTypeStockIn    EventType = "stock_in"
	EventTypeStockOut   EventType = "stock_out"
	EventTypeExpired    EventType = "expired"
	EventTypeDamaged    EventType = "damaged"
	EventTypeSold       EventType = "sold"
	EventTypeDiscounted EventType = "discounted"
)

type InventoryLog struct {
	InventoryLogId   int       `json:"inventory_log_id" db:"inventory_log_id"`
	EventType        EventType `json:"event_type"`
	ProductId        int       `json:"product_id" db:"product_id"`
	ProductCode      string    `json:"product_code" db:"product_code"`
	ProductName      string    `json:"product_name" db:"product_name"`
	ProductImagePath string    `json:"product_image_path" db:"product_image_path"`
	StockId          int       `json:"stock_id" db:"stock_id"`
	StockType        StockType `json:"stock_type" db:"stock_type"`
	StockQuantity    int       `json:"stock_quantity" db:"stock_quantity"`
	ExpiryDate       time.Time `json:"expiry_date" db:"expiry_date"`
	OriginalPrice    float64   `json:"original_price" db:"original_price"`
	DiscountedPrice  float64   `json:"discounted_price" db:"discounted_price"`
	DiscountRate     float64   `json:"discount_rate" db:"discount_rate"`
	PerformerId      int       `json:"performer_id" db:"performer_id"`
	PerformerName    string    `json:"performer_name" db:"performer_name"`
	PerformerEmail   string    `json:"performer_email" db:"performer_email"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
}

// CreateInventoryLog inserts a new inventory log and returns the created record with ID
func CreateInventoryLogWithTx(tx *sql.Tx, log *InventoryLog) (*InventoryLog, error) {
	if log == nil {
		return nil, fmt.Errorf("inventory log cannot be nil")
	}
	query := `
		INSERT INTO inventory_logs (
			event_type, product_id, product_code, product_name, product_image_path,
			stock_id, stock_type, stock_quantity, expiry_date, original_price,
			discounted_price, discount_rate, performer_id, performer_name, performer_email
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := tx.Exec(query,
		log.EventType, log.ProductId, log.ProductCode, log.ProductName, log.ProductImagePath,
		log.StockId, log.StockType, log.StockQuantity, log.ExpiryDate, log.OriginalPrice,
		log.DiscountedPrice, log.DiscountRate, log.PerformerId, log.PerformerName, log.PerformerEmail,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetInventoryLogByID(int(id))
}

// CreateInventoryLog inserts a new inventory log and returns the created record with ID
func CreateInventoryLog(log *InventoryLog) (*InventoryLog, error) {
	if log == nil {
		return nil, fmt.Errorf("inventory log cannot be nil")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		INSERT INTO inventory_logs (
			event_type, product_id, product_code, product_name, product_image_path,
			stock_id, stock_type, stock_quantity, expiry_date, original_price,
			discounted_price, discount_rate, performer_id, performer_name, performer_email
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`
	result, err := db.Exec(query,
		log.EventType, log.ProductId, log.ProductCode, log.ProductName, log.ProductImagePath,
		log.StockId, log.StockType, log.StockQuantity, log.ExpiryDate, log.OriginalPrice,
		log.DiscountedPrice, log.DiscountRate, log.PerformerId, log.PerformerName, log.PerformerEmail,
	)
	if err != nil {
		return nil, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}
	return GetInventoryLogByID(int(id))
}

// GetInventoryLogByID retrieves an inventory log by its ID
func GetInventoryLogByID(id int) (*InventoryLog, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		SELECT inventory_log_id, event_type, product_id, product_code, product_name,
			IFNULL(product_image_path, ''), stock_id, stock_type, stock_quantity,
			expiry_date, IFNULL(original_price, 0), IFNULL(discounted_price, 0),
			IFNULL(discount_rate, 0), performer_id, performer_name, performer_email, created_at
		FROM inventory_logs WHERE inventory_log_id = ?
	`
	var log InventoryLog
	err := db.QueryRow(query, id).Scan(
		&log.InventoryLogId, &log.EventType, &log.ProductId, &log.ProductCode, &log.ProductName,
		&log.ProductImagePath, &log.StockId, &log.StockType, &log.StockQuantity,
		&log.ExpiryDate, &log.OriginalPrice, &log.DiscountedPrice, &log.DiscountRate,
		&log.PerformerId, &log.PerformerName, &log.PerformerEmail, &log.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("inventory log not found")
		}
		return nil, err
	}
	return &log, nil
}

// GetInventoryLogsByProductID retrieves all inventory logs for a product, ordered by created_at desc
func GetInventoryLogsByProductID(productId int) ([]InventoryLog, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		SELECT inventory_log_id, event_type, product_id, product_code, product_name,
			IFNULL(product_image_path, ''), stock_id, stock_type, stock_quantity,
			expiry_date, IFNULL(original_price, 0), IFNULL(discounted_price, 0),
			IFNULL(discount_rate, 0), performer_id, performer_name, performer_email, created_at
		FROM inventory_logs WHERE product_id = ?
		ORDER BY created_at DESC
	`
	rows, err := db.Query(query, productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var logs []InventoryLog
	for rows.Next() {
		var log InventoryLog
		err := rows.Scan(
			&log.InventoryLogId, &log.EventType, &log.ProductId, &log.ProductCode, &log.ProductName,
			&log.ProductImagePath, &log.StockId, &log.StockType, &log.StockQuantity,
			&log.ExpiryDate, &log.OriginalPrice, &log.DiscountedPrice, &log.DiscountRate,
			&log.PerformerId, &log.PerformerName, &log.PerformerEmail, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

// GetAllInventoryLogs retrieves all inventory logs with optional filters, ordered by created_at desc
func GetAllInventoryLogs(eventType *EventType, limit, offset int) ([]InventoryLog, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		SELECT inventory_log_id, event_type, product_id, product_code, product_name,
			IFNULL(product_image_path, ''), stock_id, stock_type, stock_quantity,
			expiry_date, IFNULL(original_price, 0), IFNULL(discounted_price, 0),
			IFNULL(discount_rate, 0), performer_id, performer_name, performer_email, created_at
		FROM inventory_logs
	`
	args := []interface{}{}
	if eventType != nil {
		query += " WHERE event_type = ?"
		args = append(args, string(*eventType))
	}
	query += " ORDER BY created_at DESC"
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}
	if offset > 0 {
		query += " OFFSET ?"
		args = append(args, offset)
	}

	var rows *sql.Rows
	var err error
	if len(args) > 0 {
		rows, err = db.Query(query, args...)
	} else {
		rows, err = db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []InventoryLog
	for rows.Next() {
		var log InventoryLog
		err := rows.Scan(
			&log.InventoryLogId, &log.EventType, &log.ProductId, &log.ProductCode, &log.ProductName,
			&log.ProductImagePath, &log.StockId, &log.StockType, &log.StockQuantity,
			&log.ExpiryDate, &log.OriginalPrice, &log.DiscountedPrice, &log.DiscountRate,
			&log.PerformerId, &log.PerformerName, &log.PerformerEmail, &log.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		logs = append(logs, log)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return logs, nil
}

// UpdateInventoryLog updates an existing inventory log by ID
func UpdateInventoryLog(log *InventoryLog) error {
	if log == nil {
		return fmt.Errorf("inventory log cannot be nil")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
		UPDATE inventory_logs SET
			event_type = ?, product_id = ?, product_code = ?, product_name = ?, product_image_path = ?,
			stock_id = ?, stock_type = ?, stock_quantity = ?, expiry_date = ?,
			original_price = ?, discounted_price = ?, discount_rate = ?,
			performer_id = ?, performer_name = ?, performer_email = ?
		WHERE inventory_log_id = ?
	`
	result, err := db.Exec(query,
		log.EventType, log.ProductId, log.ProductCode, log.ProductName, log.ProductImagePath,
		log.StockId, log.StockType, log.StockQuantity, log.ExpiryDate,
		log.OriginalPrice, log.DiscountedPrice, log.DiscountRate,
		log.PerformerId, log.PerformerName, log.PerformerEmail,
		log.InventoryLogId,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("inventory log not found")
	}
	return nil
}

// DeleteInventoryLog deletes an inventory log by ID
func DeleteInventoryLog(id int) error {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `DELETE FROM inventory_logs WHERE inventory_log_id = ?`
	result, err := db.Exec(query, id)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("inventory log not found")
	}
	return nil
}
