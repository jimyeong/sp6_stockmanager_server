package service

import (
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
	respositories "github.com/jimyeongjung/owlverload_api/v1/repositories"
)

func GetStocksServiceByProductId(productId string) ([]v1.Stock, error) {
	// do something later
	return respositories.GetStocksByProductId(productId)
}
