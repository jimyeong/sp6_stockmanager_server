package controller

import (
	"time"

	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
	"github.com/jimyeongjung/owlverload_api/v1/service"
)

type FinaliseExpiredStockRequest struct {
	StockId        string       `json:"stock_id"` // pointer: nil = omitted, 0 = explicitly sent
	EventType      v1.EventType `json:"event_type"`
	StockType      v1.StockType `json:"stock_type"`
	PerformerEmail string       `json:"performer_email"`
}

type SearchInventoryRequest struct {
	SearchType string `json:"search_type"`
	Value      string `json:"value"`
}
type SearchInventoryResponse struct {
	Results    []SearchedProduct `json:"results"`
	SearchType string            `json:"search_type"`
	Value      string            `json:"value"`
	Total      int               `json:"total"`
}
type GetProductWithExpiringStocksBetweenDaysRequest struct {
	StartPeriod time.Time `json:"start_period"`
	EndPeriod   time.Time `json:"end_period"`
}
type GetProductWithExpiringStocksBetweenDaysResponse struct {
	Products  []v1.Product `json:"products"`
	StartDate string       `json:"start_date"`
	EndDate   string       `json:"end_date"`
	Total     int          `json:"total"`
	Message   string       `json:"message"`
}
type GetProductsWithStockWithDaysLeftRequest struct {
	DaysLeft int `json:"days_left"`
}
type GetProductsWithStockWithDaysLeftResponse struct {
	ExpiringProducts []v1.Product `json:"expiring_products"`
	DaysLeft         int          `json:"days_left"`
	Total            int          `json:"total"`
	Message          string       `json:"message"`
	IsExpired        bool         `json:"is_expired"`
}
type GetExpiredInventoryOlderThanDaysResponse struct {
	ProductsWithExpiredStocks []service.InventoryWithExpiredStock `json:"products_with_expired_stocks"`
	Total                     int                                 `json:"total"`
}
type GetInventoryByProductIdResponse struct {
	Products  map[string]*v1.Product `json:"products"`
	ProductId string                 `json:"product_id"`
	Total     int                    `json:"total"`
}
type GetStocksByProductIdResponse struct {
	Stocks []v1.Stock `json:"stocks"`
	Total  int        `json:"total"`
}

type CreateStockRequest struct {
	Barcode      string       `json:"barcode"`
	Code         string       `json:"code"`
	ItemID       string       `json:"item_id"`
	StockType    v1.StockType `json:"stock_type"`
	Quantity     int          `json:"quantity"`
	ExpiryDate   time.Time    `json:"expiry_date"`
	Location     string       `json:"location"`
	UserID       string       `json:"user_id"`
	Notes        string       `json:"notes"`
	DiscountRate int          `json:"discount_rate"`
}

type CreateStockResponse struct {
	Message       string     `json:"message"`
	UpdatedStocks []v1.Stock `json:"updated_stocks"`
	AddedStocks   []v1.Stock `json:"added_stocks"`
}
