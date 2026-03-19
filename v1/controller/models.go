package controller

import (
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

type FinaliseExpiredStockRequest struct {
	StockId        string       `json:"stock_id"` // pointer: nil = omitted, 0 = explicitly sent
	EventType      v1.EventType `json:"event_type"`
	StockType      v1.StockType `json:"stock_type"`
	PerformerEmail string       `json:"performer_email"`
}
