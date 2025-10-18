// apis/middleware/idempotency.go
package middleware

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"fmt"

	"github.com/jimyeongjung/owlverload_api/config"
	"github.com/jimyeongjung/owlverload_api/models"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func getRedis() *redis.Client {
	if redisClient != nil {
		return redisClient
	}
	addr := os.Getenv("REDIS_URL")

	// password := os.Getenv("REDIS_PASSWORD")
	redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   0,
	})
	return redisClient
}

var beginScript = redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if not v then
  redis.call('SET', KEYS[1], 'processing', 'PX', ARGV[1])
  return {1, ''}
elseif v == 'processing' then
  return {0, ''}
else
  return {2, v}
end
`)

var finishScript = redis.NewScript(`
local v = redis.call('GET', KEYS[1])
if v == 'processing' then
  redis.call('SET', KEYS[1], 'done:'..ARGV[1], 'PX', ARGV[2])
  return 1
else
  return 0
end
`)

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func deriveIdempotencyKey(r *http.Request, body []byte) string {
	key := r.Header.Get("Idempotency-Key")
	if key != "" {
		return key
	}
	base := r.Method + "|" + r.URL.Path + "|" + string(body)
	return sha256Hex([]byte(base))
}

func beginIdempotency(ctx context.Context, key string, processingTTLms int64) (int64, string, error) {
	c := getRedis()

	res, err := beginScript.Run(ctx, c, []string{"idem:" + key}, processingTTLms).Result()
	if err != nil {
		fmt.Println("---beginIdempotency error---", err)
		return -1, "", err
	}
	fmt.Println("---redis connection result---", res)
	vals := res.([]interface{})
	code := vals[0].(int64)
	val := ""
	if vals[1] != nil {
		val = vals[1].(string)
	}
	return code, val, nil
}

func finishIdempotency(ctx context.Context, key, ref string, doneTTLms int64) error {
	c := getRedis()
	_, err := finishScript.Run(ctx, c, []string{"idem:" + key}, ref, doneTTLms).Result()
	return err
}

func parseDoneRef(v string) string {
	const prefix = "done:"
	if strings.HasPrefix(v, prefix) {
		return strings.TrimPrefix(v, prefix)
	}
	return v
}

// ------------------------
// Idempotency Middleware
// ------------------------

type idemResponseWriter struct {
	http.ResponseWriter
	status int
}

func (w *idemResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func isMutatingMethod(m string) bool {
	return m == http.MethodPost || m == http.MethodPut || m == http.MethodDelete
}

func IdempotencyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isMutatingMethod(r.Method) {
			next.ServeHTTP(w, r)
			return
		}

		// read the body and restore it
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		r.Body = io.NopCloser(bytes.NewBuffer(body))

		idemKey := deriveIdempotencyKey(r, body)

		code, val, err := beginIdempotency(r.Context(), idemKey, config.ProcessingTTL())
		if err != nil {
			log.Printf("idempotency begin error (middleware): %v", err)
		}
		switch code {
		case 0:
			w.Header().Set("Retry-After", "3")
			// prevent duplicate request
			// return here so the handler is not called
			models.WriteServiceError(w, "Duplicate request in progress. Please retry.", false, true, http.StatusConflict)
			return
		case 2:
			// it's already responded to the client
			ref := parseDoneRef(val)
			models.WriteServiceResponse(w, "Already processed (idempotent)", map[string]any{"ref": ref}, true, true, http.StatusOK)
			return
		}

		// 핸들러 실행
		iw := &idemResponseWriter{ResponseWriter: w, status: 0}
		next.ServeHTTP(iw, r)

		if iw.status == 0 {
			iw.status = http.StatusOK
		}
		if iw.status >= 200 && iw.status < 300 {
			// if success (not 5xx) then mark as done

			ref := iw.Header().Get("Idempotency-Ref")
			if ref == "" {
				ref = sha256Hex([]byte(r.Method + "|" + r.URL.Path + "|" + string(body)))
			}

			// mark as done
			if err := finishIdempotency(r.Context(), idemKey, ref, config.DoneTTL()); err != nil {
				log.Printf("idempotency finish error (middleware): %v", err)
			}
		} else {
			// client error
			// delete the idempotency key, so the next request can proceed
			if iw.status >= 400 && iw.status < 500 {
				_ = getRedis().Del(r.Context(), "idem:"+idemKey).Err()
				models.WriteServiceError(w, "Client error. Please check your request and try again.", false, true, iw.status)
			} else {
				_ = getRedis().Del(r.Context(), "idem:"+idemKey).Err()
				models.WriteServiceError(w, "Internal server error. Please try again later.", false, true, iw.status)
			}
		}
	})
}
