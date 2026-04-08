package service

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jimyeongjung/owlverload_api/models"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
	respositories "github.com/jimyeongjung/owlverload_api/v1/repositories"
)

type ExpiringInventory struct {
	Product    v1.Product
	DaysLeft   int
	ExpiryDate time.Time
}
type InventoryWithExpiredStock struct {
	Product       v1.Product        `json:"product"`
	ExpiredStocks []v1.ExpiredStock `json:"expired_stocks"`
}

type FinaliseExpiredStockParams struct {
	StockId        string
	ProductId      int
	EventType      v1.EventType
	StockType      v1.StockType
	PerformerEmail string
}

// GetExpiryInventoryBetweenDays finds stocks that expire between startDate and endDate (inclusive),
// and returns the products that contain those stocks. Each product's Stock slice contains only
// the stocks expiring in that date range.

func GetExpiredInventoryOlderThanDays(days int) ([]InventoryWithExpiredStock, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `
	SELECT 	p.item_id,
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
			s.fkproduct_id,
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
	WHERE s.expiry_date <= Date_SUB(CURDATE(), INTERVAL ? DAY)
	ORDER BY p.item_id, s.expiry_date
	`
	rows, err := db.Query(query, days)
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
			&s.ItemId,
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
	productWithExpiredStock := make([]InventoryWithExpiredStock, 0, len(productMap))
	for _, p := range productMap {
		expiredStocks := make([]v1.ExpiredStock, 0, len(p.Stock))
		for _, s := range p.Stock {
			expiredStock := v1.ExpiredStock{
				Stock:           s,
				DaysSinceExpiry: int(time.Since(s.ExpiryDate).Hours() / 24),
			}
			expiredStocks = append(expiredStocks, expiredStock)
		}
		productWithExpiredStock = append(productWithExpiredStock, InventoryWithExpiredStock{
			Product:       *p,
			ExpiredStocks: expiredStocks,
		})
	}
	return productWithExpiredStock, nil
}

func GetInventoryWithDaysLeft(daysLeft int) ([]v1.Product, error) {
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
		WHERE DATEDIFF(s.expiry_date, CURDATE()) <= ?
	
		ORDER BY p.item_id, s.expiry_date
	`

	rows, err := db.Query(query, daysLeft)
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
func GetInventoryOfExpiringStockWithinRange(startDate, endDate string) ([]v1.Product, error) {
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

func GetInventoryByProductId(productId string) (map[string]*v1.Product, error) {
	return respositories.GetInventoryByProductId(productId)
}

func FinaliseExpiredStock(params FinaliseExpiredStockParams) error {
	db := models.GetDBInstance(models.GetDBConfig())
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stockIdInt, err := strconv.Atoi(params.StockId)
	if err != nil {
		return err
	}
	product, err := respositories.GetInventoryByStockId(stockIdInt)
	if err != nil {
		return err
	}

	performer, err := v1.GetUserByEmail(params.PerformerEmail)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	err = v1.DeleteStockByIDWithTx(tx, stockIdInt)
	if err != nil {
		return err
	}
	productIdInt, err := strconv.Atoi(product.ID)
	if err != nil {
		return err
	}
	// write log
	var stockQuantity int
	var originalPrice float64
	var discountedPrice float64
	switch params.StockType {
	case v1.StockTypeBox:
		stockQuantity = product.Stock[0].BoxNumber
		originalPrice = product.BoxPrice
		discountedPrice = product.BoxPrice * (1 - float64(product.Stock[0].DiscountRate)/100)
	case v1.StockTypePCS:
		stockQuantity = product.Stock[0].PCSNumber
		originalPrice = product.Price
		discountedPrice = product.Price * (1 - float64(product.Stock[0].DiscountRate)/100)
	case v1.StockTypeBundle:
		stockQuantity = product.Stock[0].BundleNumber
		originalPrice = product.BoxPrice
		discountedPrice = product.BoxPrice * (1 - float64(product.Stock[0].DiscountRate)/100)
	}
	_, err = v1.CreateInventoryLogWithTx(tx, &v1.InventoryLog{
		ProductId:        productIdInt,
		StockId:          stockIdInt,
		StockType:        params.StockType,
		StockQuantity:    stockQuantity,
		EventType:        params.EventType,
		ExpiryDate:       product.Stock[0].ExpiryDate,
		OriginalPrice:    originalPrice,
		DiscountedPrice:  discountedPrice,
		DiscountRate:     float64(product.Stock[0].DiscountRate),
		PerformerId:      performer.ID,
		PerformerName:    performer.DisplayName,
		PerformerEmail:   performer.Email,
		CreatedAt:        time.Now(),
		ProductCode:      product.Code,
		ProductName:      product.Name,
		ProductImagePath: product.ImagePath,
	})
	if err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}
func SearchInventoryService(searchType string, value string) ([]v1.Product, error) {
	if searchType == "" {
		return nil, fmt.Errorf("search type is required")
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return nil, fmt.Errorf("value is required")
	}

	if searchType == "code" || searchType == "barcode" || searchType == "name" {
		// sanitize value
		return respositories.GetInventoryByValue(searchType, value)
	}
	return nil, fmt.Errorf("invalid search type: %s", searchType)
}
