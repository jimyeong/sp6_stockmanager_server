package models

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the system
type RefreshToken struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	IsRevoked   bool      `json:"is_revoked"`
	UserAgent   string    `json:"user_agent,omitempty"`
	IPAddress   string    `json:"ip_address,omitempty"`
}

// GenerateRefreshToken creates a new cryptographically secure refresh token
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 64)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// Save creates a new refresh token in the database
func (rt *RefreshToken) Save() error {
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// Generate a unique ID if not provided
	if rt.ID == "" {
		rt.ID = uuid.New().String()
	}

	// Set created_at if not provided
	if rt.CreatedAt.IsZero() {
		rt.CreatedAt = time.Now()
	}

	stmt, err := db.Prepare(`
		INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at, is_revoked, user_agent, ip_address)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		rt.ID,
		rt.UserID,
		rt.Token,
		rt.ExpiresAt,
		rt.CreatedAt,
		rt.IsRevoked,
		rt.UserAgent,
		rt.IPAddress,
	)
	if err != nil {
		return fmt.Errorf("failed to save refresh token: %v", err)
	}

	return nil
}

// FindByToken retrieves a refresh token by its token string
func FindRefreshTokenByToken(token string) (*RefreshToken, error) {
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		return nil, fmt.Errorf("database connection is nil")
	}

	stmt, err := db.Prepare(`
		SELECT id, user_id, token, expires_at, created_at, last_used_at, is_revoked, user_agent, ip_address
		FROM refresh_tokens
		WHERE token = ? AND is_revoked = FALSE
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	var rt RefreshToken
	var lastUsedAt *time.Time

	err = stmt.QueryRow(token).Scan(
		&rt.ID,
		&rt.UserID,
		&rt.Token,
		&rt.ExpiresAt,
		&rt.CreatedAt,
		&lastUsedAt,
		&rt.IsRevoked,
		&rt.UserAgent,
		&rt.IPAddress,
	)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or invalid: %v", err)
	}

	rt.LastUsedAt = lastUsedAt

	return &rt, nil
}

// UpdateLastUsed updates the last_used_at timestamp
func (rt *RefreshToken) UpdateLastUsed() error {
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	now := time.Now()
	stmt, err := db.Prepare(`
		UPDATE refresh_tokens
		SET last_used_at = ?
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(now, rt.ID)
	if err != nil {
		return fmt.Errorf("failed to update last used timestamp: %v", err)
	}

	rt.LastUsedAt = &now
	return nil
}

// Revoke marks the refresh token as revoked
func (rt *RefreshToken) Revoke() error {
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	stmt, err := db.Prepare(`
		UPDATE refresh_tokens
		SET is_revoked = TRUE
		WHERE id = ?
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(rt.ID)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %v", err)
	}

	rt.IsRevoked = true
	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a specific user
func RevokeAllUserTokens(userID string) error {
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	stmt, err := db.Prepare(`
		UPDATE refresh_tokens
		SET is_revoked = TRUE
		WHERE user_id = ? AND is_revoked = FALSE
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %v", err)
	}

	return nil
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (rt *RefreshToken) IsValid() bool {
	return !rt.IsRevoked && time.Now().Before(rt.ExpiresAt)
}

// CleanupExpiredTokens removes all expired refresh tokens from the database
func CleanupExpiredTokens() error {
	db := GetDBInstance(GetDBConfig())
	if db == nil {
		return fmt.Errorf("database connection is nil")
	}

	stmt, err := db.Prepare(`
		DELETE FROM refresh_tokens
		WHERE expires_at < ? OR is_revoked = TRUE
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %v", err)
	}

	return nil
}
