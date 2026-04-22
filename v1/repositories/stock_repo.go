package respositories

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jimyeongjung/owlverload_api/models"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

const insertStockQuery = `
	INSERT INTO stocks (fkproduct_id, stock_type, box_number, pcs_number, bundle_number, expiry_date, location, registering_person, notes, discount_rate)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`

const getStocksByProductIdQuery = `
	SELECT stock_id, fkproduct_id, stock_type, box_number, pcs_number, bundle_number, expiry_date,
		IFNULL(location, ''), IFNULL(registering_person, ''), IFNULL(notes, ''), IFNULL(discount_rate, 0), created_at
	FROM stocks WHERE fkproduct_id = ?
`

func GetStocksByProductId(productId string) ([]v1.Stock, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	rows, err := db.Query(getStocksByProductIdQuery, productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStocks(rows)
}

func GetStocksByProductIdWithTx(tx *sql.Tx, productId string) ([]v1.Stock, error) {
	rows, err := tx.Query(getStocksByProductIdQuery, productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanStocks(rows)
}

func scanStocks(rows *sql.Rows) ([]v1.Stock, error) {
	var stocks []v1.Stock
	for rows.Next() {
		var s v1.Stock
		err := rows.Scan(
			&s.StockId,
			&s.ItemId,
			&s.StockType,
			&s.BoxNumber,
			&s.PCSNumber,
			&s.BundleNumber,
			&s.ExpiryDate,
			&s.Location,
			&s.RegisteringPerson,
			&s.Notes,
			&s.DiscountRate,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return stocks, nil
}

func CreateStock(stock v1.Stock) (int64, error) {
	if stock.ItemId == "" {
		return 0, errors.New("product id (fkproduct_id) is required")
	}

	db := models.GetDBInstance(models.GetDBConfig())
	result, err := db.Exec(insertStockQuery,
		stock.ItemId,
		stock.StockType,
		stock.BoxNumber,
		stock.PCSNumber,
		stock.BundleNumber,
		stock.ExpiryDate,
		stock.Location,
		stock.RegisteringPerson,
		stock.Notes,
		stock.DiscountRate,
	)
	// check result
	if err != nil {
		return 0, fmt.Errorf("failed to insert stock: %w", err)
	}
	// return the last insert id
	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

func CreateStockWithTx(tx *sql.Tx, stock v1.Stock) error {
	if stock.ItemId == "" {
		return errors.New("product id (fkproduct_id) is required")
	}
	_, err := tx.Exec(insertStockQuery,
		stock.ItemId,
		stock.StockType,
		stock.BoxNumber,
		stock.PCSNumber,
		stock.BundleNumber,
		stock.ExpiryDate,
		stock.Location,
		stock.RegisteringPerson,
		stock.Notes,
		stock.DiscountRate,
	)
	return err
}

func DeleteStockById(stockId string) error {
	db := models.GetDBInstance(models.GetDBConfig())
	_, err := db.Exec("DELETE FROM stocks WHERE stock_id = ?", stockId)
	if err != nil {
		return err
	}
	return nil
}

func UpdateStockById(stockId string, stock v1.Stock) error {
	db := models.GetDBInstance(models.GetDBConfig())
	_, err := db.Exec(`
		UPDATE stocks SET expiry_date = ?, 
		location = ?, 
		registering_person = ?, 
		notes = ?, 
		discount_rate = ?,
		box_number = ?,
		pcs_number = ?,
		stock_type = ?
		WHERE stock_id = ?
	`, stock.ExpiryDate,
		stock.Location,
		stock.RegisteringPerson,
		stock.Notes,
		stock.DiscountRate,
		stock.BoxNumber,
		stock.PCSNumber,
		stock.StockType,
		stockId)
	if err != nil {
		return err
	}
	return nil
}
