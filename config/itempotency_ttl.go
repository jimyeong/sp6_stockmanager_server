package config

import (
	"os"
	"strconv"
	"sync"
)

const (
	// 기본값 (운영에서는 환경변수로 덮어쓰기)
	defaultProcessingTTLms int64 = 30_000     // 30s
	defaultDoneTTLms       int64 = 86_400_000 // 24h
)

var (
	ttlOnce sync.Once
	procTTL int64 = defaultProcessingTTLms
	doneTTL int64 = defaultDoneTTLms
)

// ENV
//
//	IDEM_PROCESSING_TTL_MS  예: 30000 (30초)
//	IDEM_DONE_TTL_MS        예: 86400000 (24시간)
func initTTLs() {
	if v := os.Getenv("IDEM_PROCESSING_TTL_MS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			procTTL = n
		}
	}
	if v := os.Getenv("IDEM_DONE_TTL_MS"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 {
			doneTTL = n
		}
	}
}

// 임시 처리 락 TTL(ms)
func ProcessingTTL() int64 {
	ttlOnce.Do(initTTLs)
	return procTTL
}

// 완료 상태 TTL(ms)
func DoneTTL() int64 {
	ttlOnce.Do(initTTLs)
	return doneTTL
}
