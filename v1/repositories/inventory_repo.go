package respositories

import (
	"github.com/jimyeongjung/owlverload_api/models"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

func GetInventoryByProductId(productId string) (map[string]*v1.Product, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
	SELECT 
		p.item_id,
		p.code,
		p.barcode,
		p.box_barcode,
		p.price,
		p.box_price,
		p.name,
		p.type,
		p.available_for_order,
		p.image_path,
		p.created_at,
		p.name_jpn,
		p.name_chn,
		p.name_kor,
		p.name_eng,
		s.stock_id,
		s.expiry_date,
		s.stock_type,
		s.box_number,
		s.pcs_number,
		s.bundle_number,
		s.location,
		s.registering_person,
		s.notes,
		s.discount_rate,
		s.created_at,
	FROM items as p
	LEFT JOIN stocks as s ON p.item_id = s.fkproduct_id
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
	return pMap, nil
}

func GetInventoryByStockId(stockId int) (*v1.Product, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
	SELECT 
		p.item_id,
		p.code,
		p.barcode,
		p.box_barcode,
		p.price,
		p.box_price,
		p.name,
		p.type,
		p.available_for_order,
		p.image_path,
		p.created_at,
		p.name_jpn,
		p.name_chn,
		p.name_kor,
		p.name_eng,
		s.stock_id,
		s.expiry_date,
		s.stock_type,
		s.box_number,
		s.pcs_number,
		s.bundle_number,
		s.location,
		s.registering_person,
		s.notes,
		s.discount_rate,
		s.created_at,
	FROM items as p
	LEFT JOIN stocks as s ON p.item_id = s.fkproduct_id
	WHERE s.stock_id = ?
	`
	rows, err := db.Query(query, stockId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var p v1.Product
	var s v1.Stock
	err = rows.Scan(
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
	return &p, nil
}
