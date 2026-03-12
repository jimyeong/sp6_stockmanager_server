package v1

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jimyeongjung/owlverload_api/models"
)

type Product struct {
	ID                string    `json:"product_id" db:"item_id"`
	Code              string    `json:"code" db:"code"`
	BarCode           string    `json:"barcode" db:"barcode"`
	BoxBarcode        string    `json:"box_barcode" db:"box_barcode"`
	Price             float64   `json:"price" db:"price"`
	BoxPrice          float64   `json:"box_price" db:"box_price"`
	Name              string    `json:"name" db:"name"`
	Type              string    `json:"type" db:"type"`
	AvailableForOrder int       `json:"availableForOrder" db:"available_for_order"`
	ImagePath         string    `json:"image_path" db:"image_path"`
	CreatedAt         time.Time `json:"createdAt,omitempty" db:"created_at"`
	NameJpn           string    `json:"name_jpn" db:"name_jpn"`
	NameChn           string    `json:"name_chn" db:"name_chn"`
	NameKor           string    `json:"name_kor" db:"name_kor"`
	NameEng           string    `json:"name_eng" db:"name_eng"`
	Stock             []Stock   `json:"stock" db:"stock"`
	Tag               []Tag     `json:"tag" db:"tag"`
	Ingredients       string    `json:"ingredients" db:"ingredients"`
	IsBeefContained   bool      `json:"is_beef_contained" db:"is_beef_contained"`
	IsPorkContained   bool      `json:"is_pork_contained" db:"is_pork_contained"`
	IsHalal           bool      `json:"is_halal" db:"is_halal"`
	IsPlantBased      bool      `json:"is_plant_based" db:"is_plant_based"`
	Reasoning         string    `json:"reasoning" db:"reasoning"`
}

// GetProductByID retrieves a product by its ID (item_id)
func GetProductByID(id string) (Product, error) {
	if id == "" {
		return Product{}, fmt.Errorf("empty product ID")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	var p Product
	query := `SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(box_barcode, ''), IFNULL(price, 0), IFNULL(box_price, 0), IFNULL(name, ''), IFNULL(type, ''),
		IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at,
		IFNULL(name_jpn, ''), IFNULL(name_chn, ''), IFNULL(name_kor, ''), IFNULL(name_eng, '')
		FROM items WHERE item_id = ?`
	err := db.QueryRow(query, id).Scan(
		&p.ID, &p.Code, &p.BarCode, &p.BoxBarcode, &p.Price, &p.BoxPrice, &p.Name, &p.Type,
		&p.AvailableForOrder, &p.ImagePath, &p.CreatedAt,
		&p.NameJpn, &p.NameChn, &p.NameKor, &p.NameEng,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Product{}, errors.New("product not found")
		}
		return Product{}, err
	}
	return p, nil
}

// GetProductByIDWithStockAndTags retrieves a product by ID with its stock and tags
func GetProductByIDWithStockAndTags(id string) (Product, error) {
	p, err := GetProductByID(id)
	if err != nil {
		return Product{}, err
	}
	stocks, err := GetStocksByProductID(id)
	if err != nil {
		p.Stock = []Stock{}
	} else {
		p.Stock = stocks
	}
	tags, err := GetTagsForProduct(id)
	if err != nil {
		p.Tag = []Tag{}
	} else {
		p.Tag = tags
	}
	return p, nil
}

// GetProductByBarcode retrieves a product by barcode or box_barcode
func GetProductByBarcode(barcode string) (Product, error) {
	if barcode == "" {
		return Product{}, fmt.Errorf("empty barcode")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	var p Product
	query := `SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(box_barcode, ''), IFNULL(price, 0), IFNULL(box_price, 0), IFNULL(name, ''), IFNULL(type, ''),
		IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at,
		IFNULL(name_jpn, ''), IFNULL(name_chn, ''), IFNULL(name_kor, ''), IFNULL(name_eng, '')
		FROM items WHERE barcode = ? OR box_barcode = ?`
	err := db.QueryRow(query, barcode, barcode).Scan(
		&p.ID, &p.Code, &p.BarCode, &p.BoxBarcode, &p.Price, &p.BoxPrice, &p.Name, &p.Type,
		&p.AvailableForOrder, &p.ImagePath, &p.CreatedAt,
		&p.NameJpn, &p.NameChn, &p.NameKor, &p.NameEng,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Product{}, errors.New("product not found")
		}
		return Product{}, err
	}
	return p, nil
}

// GetProductByCode retrieves a product by its code
func GetProductByCode(code string) (Product, error) {
	if code == "" {
		return Product{}, fmt.Errorf("empty product code")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	var p Product
	query := `SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(box_barcode, ''), IFNULL(price, 0), IFNULL(box_price, 0), IFNULL(name, ''), IFNULL(type, ''),
		IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at,
		IFNULL(name_jpn, ''), IFNULL(name_chn, ''), IFNULL(name_kor, ''), IFNULL(name_eng, '')
		FROM items WHERE code = ?`
	err := db.QueryRow(query, code).Scan(
		&p.ID, &p.Code, &p.BarCode, &p.BoxBarcode, &p.Price, &p.BoxPrice, &p.Name, &p.Type,
		&p.AvailableForOrder, &p.ImagePath, &p.CreatedAt,
		&p.NameJpn, &p.NameChn, &p.NameKor, &p.NameEng,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return Product{}, errors.New("product not found")
		}
		return Product{}, err
	}
	return p, nil
}

// GetAllProducts retrieves all products from the database
func GetAllProducts() ([]Product, error) {
	db := models.GetDBInstance(models.GetDBConfig())
	query := `SELECT item_id, IFNULL(code, ''), IFNULL(barcode, ''), IFNULL(box_barcode, ''), IFNULL(price, 0), IFNULL(box_price, 0), IFNULL(name, ''), IFNULL(type, ''),
		IFNULL(available_for_order, 0), IFNULL(image_path, ''), created_at,
		IFNULL(name_jpn, ''), IFNULL(name_chn, ''), IFNULL(name_kor, ''), IFNULL(name_eng, '')
		FROM items`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var products []Product
	for rows.Next() {
		var p Product
		err := rows.Scan(
			&p.ID, &p.Code, &p.BarCode, &p.BoxBarcode, &p.Price, &p.BoxPrice, &p.Name, &p.Type,
			&p.AvailableForOrder, &p.ImagePath, &p.CreatedAt,
			&p.NameJpn, &p.NameChn, &p.NameKor, &p.NameEng,
		)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return products, nil
}
