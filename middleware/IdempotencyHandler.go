package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/jimyeongjung/owlverload_api/models"
)

type BusinessFunc func(ctx context.Context, rawBody []byte) (int, any, error)

func IdempotencyHandler(store *models.IdemStore, fn BusinessFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		idemKey := r.Header.Get("Idempotency-Key")
		if idemKey == "" {
			http.Error(w, "Missing Idempotency-Key", http.StatusBadRequest)
			return
		}

		raw, _ := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(raw))
		hash := Sha256sum(raw)

		state, statusStr, cachedBody, err := store.Acquire(ctx, idemKey, hash)
		if err != nil {
			http.Error(w, "Failed to acquire idempotency key", http.StatusInternalServerError)
			return
		}

		switch state {
		case "MISMATCH":
			http.Error(w, "Requeset body mismatch for same Idempotency-Key", http.StatusConflict)
			return
		case "DONE":
			st, _ := strconv.Atoi(statusStr)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(st)
			w.Write([]byte(cachedBody))
			return
		case "PROCESSING":
			w.Header().Set("Retry-After", "3")
			http.Error(w, "Processing", http.StatusConflict)
			return
		case "LOCKED":
			status, respObj, bizErr := fn(ctx, raw)
			respJSON, _ := json.Marshal(respObj)

			if bizErr == nil && status < 500 {
				_ = store.Complete(ctx, idemKey, status, string(respJSON))
			} else {
				_ = store.Rdb.Del(ctx, "idem: "+idemKey).Err()
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			w.Write(respJSON)
			return
		default:
			http.Error(w, "Failed to acquire idempotency key", http.StatusInternalServerError)
			return
		}

	}

}

func Sha256sum(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}
