package controller

import (
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

type FinaliseExpiredStockRequest struct {
	StockId     int
	EventType   v1.EventType
	StockType   v1.StockType
	PerformerId int
}
