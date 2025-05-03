package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/jimyeongjung/owlverload_api/models"
)

// TokenValidator is a function type that validates a token
type TokenValidator func(token string) (bool, string, error)

// AuthenticationConfig holds configuration for the authentication middleware
type AuthenticationConfig struct {
	ValidateToken     TokenValidator
	ExcludedPaths     []string
	TokenErrorMessage string
}

// DefaultTokenValidator is a simple token validator for testing
func DefaultTokenValidator(token string) (bool, string, error) {
	// In production, replace this with real token validation logic
	if token == "" {
		return false, "", fmt.Errorf("token is empty")
	}
	// For testing, accept any non-empty token
	return true, "test-user-id", nil
}

// NewAuthentication creates a new authentication middleware
func NewAuthentication(config AuthenticationConfig) func(http.Handler) http.Handler {
	fmt.Println("--- NewAuthentication --- ")
	if config.ValidateToken == nil {
		config.ValidateToken = DefaultTokenValidator
	}

	if config.TokenErrorMessage == "" {
		config.TokenErrorMessage = "Authentication required"
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if path is excluded from authentication
			for _, path := range config.ExcludedPaths {
				if path == r.URL.Path {
					next.ServeHTTP(w, r)
					return
				}
			}

			// Extract Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				models.WriteServiceError(w, config.TokenErrorMessage, false, false, http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				models.WriteServiceError(w, "Invalid authorization format", false, false, http.StatusUnauthorized)
				return
			}

			token := parts[1]

			// Validate token
			valid, userID, err := config.ValidateToken(token)
			if err != nil || !valid {
				errorMessage := config.TokenErrorMessage
				if err != nil {
					errorMessage = err.Error()
				}
				models.WriteServiceError(w, errorMessage, false, false, http.StatusUnauthorized)
				return
			}

			// Add user ID to request context
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserIDFromContext extracts the userID from the request context
func GetUserIDFromContext(r *http.Request) string {
	userID, ok := r.Context().Value("userID").(string)
	if !ok {
		return ""
	}
	return userID
}
