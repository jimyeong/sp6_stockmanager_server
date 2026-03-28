package respositories

import (
	"database/sql"

	v1 "github.com/jimyeongjung/owlverload_api/v1/models"
)

func SaveWithTx(tx *sql.Tx, params v1.IdempotencyRecord) error {
	query := `INSERT INTO idempotency_records (idempotency_key, resource_id, response_body, response_code, endpoint, request_hash, updated_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := tx.Exec(
		query,
		params.IdempotencyKey,
		params.ResourceId,
		params.ResponseBody,
		params.ResponseCode,
		params.Endpoint,
		params.RequestHash,
		params.UpdatedAt,
		params.CreatedAt,
	)
	return err
}

func SaveRecord(record v1.IdempotencyRecord, db *sql.DB) error {

	query := `INSERT INTO idempotency_records (idempotency_key, resource_id, response_body, response_code, endpoint, request_hash, updated_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	_, err := db.Exec(
		query,
		record.IdempotencyKey,
		record.ResourceId,
		record.ResponseBody,
		record.ResponseCode,
		record.Endpoint,
		record.RequestHash,
		record.UpdatedAt,
		record.CreatedAt,
	)
	return err
}
