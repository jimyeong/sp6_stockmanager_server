package respositories

import (
	"database/sql"
	"errors"

	"github.com/jimyeongjung/owlverload_api/models"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

const getProductByIdQuery = `
	SELECT 
		p.item_id,
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
		IFNULL(p.name_eng, '')
	FROM items AS p
	WHERE p.item_id = ?
`

const updateProductQuery = `
	UPDATE items 
	SET code = ?, barcode = ?, box_barcode = ?, price = ?, box_price = ?,
		name = ?, type = ?, name_jpn = ?, name_chn = ?, name_kor = ?, name_eng = ?,
		available_for_order = ?, image_path = ?
	WHERE item_id = ?
`

func UpdateProductRepo(product v1.Product) (v1.Product, error) {
	if product.ID == "" {
		return v1.Product{}, errors.New("product ID is required")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	_, err := db.Exec(updateProductQuery,
		product.Code, product.BarCode, product.BoxBarcode, product.Price, product.BoxPrice,
		product.Name, product.Type, product.NameJpn, product.NameChn, product.NameKor, product.NameEng,
		product.AvailableForOrder, product.ImagePath,
		product.ID,
	)
	return product, err
}

func UpdateProductWithTxRepo(tx *sql.Tx, product v1.Product) (v1.Product, error) {
	if product.ID == "" {
		return v1.Product{}, errors.New("product ID is required")
	}
	_, err := tx.Exec(updateProductQuery,
		product.Code, product.BarCode, product.BoxBarcode, product.Price, product.BoxPrice,
		product.Name, product.Type, product.NameJpn, product.NameChn, product.NameKor, product.NameEng,
		product.AvailableForOrder, product.ImagePath,
		product.ID,
	)
	return product, err
}

func DeleteProductRepo(productId string) error {
	if productId == "" {
		return errors.New("product ID is required")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := deleteProductWithTx(tx, productId); err != nil {
		return err
	}
	return tx.Commit()
}

func DeleteProductWithTxRepo(tx *sql.Tx, productId string) error {
	if productId == "" {
		return errors.New("product ID is required")
	}
	return deleteProductWithTx(tx, productId)
}

func deleteProductWithTx(tx *sql.Tx, productId string) error {
	// Delete stocks first (foreign key to items)
	if _, err := tx.Exec("DELETE FROM stocks WHERE fkproduct_id = ?", productId); err != nil {
		return err
	}
	// Delete item_tags (foreign key to items)
	if _, err := tx.Exec("DELETE FROM item_tags WHERE item_id = ?", productId); err != nil {
		return err
	}
	// Delete the item
	result, err := tx.Exec("DELETE FROM items WHERE item_id = ?", productId)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return errors.New("product not found")
	}
	return nil
}

func GetProductByIdRepo(productId string) (v1.Product, error) {
	if productId == "" {
		return v1.Product{}, errors.New("product ID is required")
	}
	db := models.GetDBInstance(models.GetDBConfig())
	return scanProduct(db.QueryRow(getProductByIdQuery, productId))
}

func GetProductByIdWithTxRepo(tx *sql.Tx, productId string) (v1.Product, error) {
	if productId == "" {
		return v1.Product{}, errors.New("product ID is required")
	}
	return scanProduct(tx.QueryRow(getProductByIdQuery, productId))
}

func scanProduct(row *sql.Row) (v1.Product, error) {
	var p v1.Product
	err := row.Scan(
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
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return v1.Product{}, errors.New("product not found")
		}
		return v1.Product{}, err
	}
	return p, nil
}
