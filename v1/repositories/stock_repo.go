package respositories

import (
	"database/sql"

	"github.com/jimyeongjung/owlverload_api/models"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

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
