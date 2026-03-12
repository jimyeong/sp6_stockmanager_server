package service

import (
	"time"

	"github.com/jimyeongjung/owlverload_api/models"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

type ExpiringProductStock struct {
	Product    v1.Product
	DaysLeft   int
	ExpiryDate time.Time
}

func GetInventoryByProductId(productId string) (v1.Product, error) {
	product, err := v1.GetProductByID(productId)
	if err != nil {
		return v1.Product{}, err
	}
	return product, nil
}
func GetExpiryInventoryByLessThanDays(days int) (v1.Product, error) {
	return v1.Product{}, nil // TODO: implement
}

// GetExpiryInventoryBetweenDays finds stocks that expire between startDate and endDate (inclusive),
// and returns the products that contain those stocks. Each product's Stock slice contains only
// the stocks expiring in that date range.

func GetExpiryInventoryBetweenDays(startDate, endDate time.Time) ([]v1.Product, error) {
	stocks, err := v1.GetStocksByExpiryDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	if len(stocks) == 0 {
		return []v1.Product{}, nil
	}
	// Group stocks by product ID
	productStocks := make(map[string][]v1.Stock)
	for _, s := range stocks {
		productStocks[s.ItemId] = append(productStocks[s.ItemId], s)
	}
	// Fetch each product and attach its expiring stocks
	var products []v1.Product
	for productID, expiringStocks := range productStocks {
		product, err := v1.GetProductByIDWithStockAndTags(productID)
		if err != nil {
			continue // skip products we can't fetch
		}
		product.Stock = expiringStocks
		products = append(products, product)
	}
	return products, nil
}

func GetProductsWithExpiringStockWithinRange(startDate, endDate string) ([]v1.Product, error) {
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
			s.stock_id,
			s.expiry_date,
			IFNULL(s.stock_type, ''),
			IFNULL(s.box_number, 0),
			IFNULL(s.pcs_number, 0),
			IFNULL(s.bundle_number, 0),
			IFNULL(s.location, ''),
			IFNULL(s.registering_person, ''),
			IFNULL(s.notes, ''),
			IFNULL(s.discount_rate, 0),
			s.created_at
		FROM items p
		JOIN stocks s ON p.item_id = s.fkproduct_id
		WHERE s.expiry_date >= ? AND s.expiry_date <= ?
		ORDER BY p.item_id, s.expiry_date
	`

	rows, err := db.Query(query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	productMap := make(map[string]*v1.Product)

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

		existing, ok := productMap[p.ID]
		if !ok {
			p.Stock = []v1.Stock{}
			productMap[p.ID] = &p
			existing = &p
		}

		existing.Stock = append(existing.Stock, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	products := make([]v1.Product, 0, len(productMap))
	for _, p := range productMap {
		products = append(products, *p)
	}

	return products, nil
}
