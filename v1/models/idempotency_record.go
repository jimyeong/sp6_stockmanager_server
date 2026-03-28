package v1

import (
	"time"
)

type IdempotencyRecord struct {
	IdempotencyKey string
	ResourceId     string
	ResponseBody   []byte
	ResponseCode   int
	Endpoint       string
	RequestHash    string
	UpdatedAt      time.Time
	CreatedAt      time.Time
}
