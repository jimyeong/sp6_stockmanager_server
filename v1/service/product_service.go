package service

import (
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
	respositories "github.com/jimyeongjung/owlverload_api/v1/repositories"
)

// EditProductService validates and updates a product.
func UpdateProductService(product v1.Product) error {
	// validate product if needed

	_, err := respositories.UpdateProductRepo(product)
	return err
}
