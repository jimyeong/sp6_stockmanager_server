package respositories

import (
	"database/sql"

	"github.com/jimyeongjung/owlverload_api/models"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

func GetInventoryByProductId(productId string) (map[string]*v1.Product, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
	SELECT 
		p.item_id,
		p.code,
		IFNULL(p.barcode, ''),
		IFNULL(p.box_barcode, ''),
		IFNULL(p.price, 0),
		IFNULL(p.box_price, 0),
		IFNULL(p.name, ''),
		IFNULL(p.type, ''),
		IFNULL(p.available_for_order, 0),
		IFNULL(p.image_path, ''),
		p.created_at,
		IFNULL(p.name_jpn, ''),
		IFNULL(p.name_chn, ''),
		IFNULL(p.name_kor, ''),
		IFNULL(p.name_eng, ''),
		IFNULL(s.stock_id, 0),
		expiry_date,
		IFNULL(s.stock_type, ''),
		IFNULL(s.box_number, 0),
		IFNULL(s.pcs_number, 0),
		IFNULL(s.bundle_number, 0),
		IFNULL(s.location, ''),
		IFNULL(s.registering_person, ''),
		IFNULL(s.notes, ''),
		IFNULL(s.discount_rate, 0),
		s.created_at
	FROM items as p
	JOIN stocks as s ON p.item_id = s.fkproduct_id
	WHERE p.item_id = ?
	`
	rows, err := db.Query(query, productId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pMap := make(map[string]*v1.Product)
	for rows.Next() {
		var p v1.Product
		var s v1.Stock
		err := rows.Scan(
			&p.ID,
			&p.Code,
			&p.BarCode,
			&p.BoxBarcode,
			&p.Price,
			&p.BoxPrice,
			&p.Name,
			&p.Type,
			&p.AvailableForOrder,
			&p.ImagePath,
			&p.CreatedAt,
			&p.NameJpn,
			&p.NameChn,
			&p.NameKor,
			&p.NameEng,
			&s.StockId,
			&s.ExpiryDate,
			&s.StockType,
			&s.BoxNumber,
			&s.PCSNumber,
			&s.BundleNumber,
			&s.Location,
			&s.RegisteringPerson,
			&s.Notes,
			&s.DiscountRate,
			&s.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		existing, ok := pMap[p.ID]
		if !ok {
			p.Stock = []v1.Stock{}
			pMap[p.ID] = &p
			existing = &p
		}
		existing.Stock = append(existing.Stock, s)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return pMap, nil
}

func GetInventoryByStockId(stockId int) (*v1.Product, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
	SELECT 
		IFNULL(p.item_id, 0),
		IFNULL(p.code, ''),
		IFNULL(p.barcode, ''),
		IFNULL(p.box_barcode, ''),
		IFNULL(p.price, 0),
		IFNULL(p.box_price, 0),
		IFNULL(p.name, ''),
		IFNULL(p.type, ''),
		IFNULL(p.available_for_order, 0),
		IFNULL(p.image_path, ''),
		p.created_at,
		IFNULL(p.name_jpn, ''),
		IFNULL(p.name_chn, ''),
		IFNULL(p.name_kor, ''),
		IFNULL(p.name_eng, ''),
		IFNULL(s.stock_id, 0),
		s.expiry_date,
		IFNULL(s.stock_type, ''),
		IFNULL(s.box_number, 0),
		IFNULL(s.pcs_number, 0),
		IFNULL(s.bundle_number, 0),
		IFNULL(s.location, ''),
		IFNULL(s.registering_person, ''),
		IFNULL(s.notes, ''),
		IFNULL(s.discount_rate, 0),
		IFNULL(s.created_at, p.created_at)
	FROM items as p
	JOIN stocks as s ON p.item_id = s.fkproduct_id
	WHERE s.stock_id = ?
	`
	var p v1.Product
	var s v1.Stock
	err := db.QueryRow(query, stockId).Scan(
		&p.ID,
		&p.Code,
		&p.BarCode,
		&p.BoxBarcode,
		&p.Price,
		&p.BoxPrice,
		&p.Name,
		&p.Type,
		&p.AvailableForOrder,
		&p.ImagePath,
		&p.CreatedAt,
		&p.NameJpn,
		&p.NameChn,
		&p.NameKor,
		&p.NameEng,
		&s.StockId,
		&s.ExpiryDate,
		&s.StockType,
		&s.BoxNumber,
		&s.PCSNumber,
		&s.BundleNumber,
		&s.Location,
		&s.RegisteringPerson,
		&s.Notes,
		&s.DiscountRate,
		&s.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p.Stock = []v1.Stock{s}
	return &p, nil
}
