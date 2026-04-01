package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"strconv"

	"github.com/jimyeongjung/owlverload_api/models"
	appredis "github.com/jimyeongjung/owlverload_api/v1/internal/redis"
	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
	respositories "github.com/jimyeongjung/owlverload_api/v1/repositories"
)

func GetStocksServiceByProductId(productId string) ([]v1.Stock, error) {
	// do something later
	return respositories.GetStocksByProductId(productId)
}

// hashStockPayload returns a stable SHA-256 hex digest of the fields that identify the create-stock request.
// It runs after quantity is mapped into box/pcs fields so the same logical request hashes the same way.
func hashStockPayload(stock v1.Stock) (string, error) {
	p := struct {
		ItemID            string    `json:"item_id"`
		StockType         string    `json:"stock_type"`
		PCSNumber         int       `json:"pcs_number"`
		BoxNumber         int       `json:"box_number"`
		BundleNumber      int       `json:"bundle_number"`
		ExpiryDate        time.Time `json:"expiry_date"`
		Location          string    `json:"location"`
		RegisteringPerson string    `json:"registering_person"`
		Notes             string    `json:"notes"`
		DiscountRate      int       `json:"discount_rate"`
	}{
		ItemID:            stock.ItemId,
		StockType:         string(stock.StockType),
		PCSNumber:         stock.PCSNumber,
		BoxNumber:         stock.BoxNumber,
		BundleNumber:      stock.BundleNumber,
		ExpiryDate:        stock.ExpiryDate,
		Location:          stock.Location,
		RegisteringPerson: stock.RegisteringPerson,
		Notes:             stock.Notes,
		DiscountRate:      stock.DiscountRate,
	}
	b, err := json.Marshal(p)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func CreateStockService(ctx context.Context, stock v1.Stock, idempotencyKey string) error {
	stock.StockType = v1.StockType(stock.StockType)

	if idempotencyKey == "" {
		return errors.New("idempotency key is required")
	}

	switch stock.StockType {
	case v1.StockTypePCS:
		if stock.Quantity <= 0 {
			return errors.New("quantity must be greater than 0")
		}
		stock.PCSNumber = stock.Quantity
	case v1.StockTypeBox:
		if stock.Quantity <= 0 {
			return errors.New("quantity must be greater than 0")
		}
		stock.BoxNumber = stock.Quantity
	default:
		return errors.New("invalid stock type")
	}

	requestHash, err := hashStockPayload(stock)
	if err != nil {
		return errors.New("failed to hash stock object")
	}

	opts, err := appredis.LoadOptionsFromEnv()
	if err != nil {
		return err
	}
	redisClient, err := appredis.GetRedisClient(ctx, opts)
	if err != nil {
		return err
	}
	redisKey := "idem:create_stock:" + idempotencyKey

	ok, err := redisClient.SetNX(ctx, redisKey, "processing", time.Hour).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("duplicate request")
	}

	lastInsertId, err := respositories.CreateStock(stock)
	if err != nil {
		cmd := redisClient.Del(ctx, redisKey)
		if cmd.Err() != nil {
			return cmd.Err()
		}
		return err
	}

	err = redisClient.Set(ctx, redisKey, "completed", time.Hour).Err()
	if err != nil {
		return err
	}

	record := v1.IdempotencyRecord{
		IdempotencyKey: idempotencyKey,
		ResourceId:     strconv.FormatInt(lastInsertId, 10),
		ResponseBody:   []byte("stock_created"),
		ResponseCode:   200,
		Endpoint:       "/api/v1/stocks/create",
		RequestHash:    requestHash,
		UpdatedAt:      time.Now(),
		CreatedAt:      time.Now(),
	}
	err = respositories.SaveRecord(record, models.GetDBInstance(models.GetDBConfig()))
	if err != nil {
		// best-effort only; do not fail the stock creation after DB insert succeeded
	}

	return nil
}

func DeleteStockByIdService(ctx context.Context, stockId string) error {
	if stockId == "" {
		return errors.New("stockId is required")
	}
	return respositories.DeleteStockById(stockId)
}
