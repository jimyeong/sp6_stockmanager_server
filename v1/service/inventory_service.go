package service

import (
	"time"

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
func GetExpiryInventoryByLessThanDays(days int) (v1.Product, error) {}
func GetExpiryInventoryBetweenDays(startDate time.Time, endDate time.Time) (v1.Product, error) {

}
