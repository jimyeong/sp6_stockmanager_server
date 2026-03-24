package apis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"firebase.google.com/go/auth"
	"github.com/jimyeongjung/owlverload_api/models"
)

// RefreshTokenRequest defines the request structure for token refresh
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// RefreshTokenResponse defines the response structure for token refresh
type RefreshTokenResponse struct {
	CustomToken  string `json:"custom_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

// HandleRefreshToken handles POST requests to refresh Firebase ID tokens using a refresh token
func HandleRefreshToken(w http.ResponseWriter, r *http.Request, firebaseClient *auth.Client) {
	fmt.Println("---HandleRefreshToken---")

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, false, http.StatusBadRequest)
		return
	}

	var refreshReq RefreshTokenRequest
	err = json.Unmarshal(body, &refreshReq)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, false, http.StatusBadRequest)
		return
	}

	if refreshReq.RefreshToken == "" {
		models.WriteServiceError(w, "Refresh token is required", false, false, http.StatusBadRequest)
		return
	}

	// Find the refresh token in the database
	storedToken, err := models.FindRefreshTokenByToken(refreshReq.RefreshToken)
	if err != nil {
		fmt.Println("Refresh token not found:", err)
		models.WriteServiceError(w, "Invalid or expired refresh token", false, false, http.StatusUnauthorized)
		return
	}

	// Check if the token is valid
	if !storedToken.IsValid() {
		fmt.Println("Refresh token is invalid or expired")
		models.WriteServiceError(w, "Refresh token has expired or been revoked", false, false, http.StatusUnauthorized)
		return
	}

	// Update the last used timestamp
	err = storedToken.UpdateLastUsed()
	if err != nil {
		fmt.Println("Failed to update last used timestamp:", err)
		// Continue anyway - this is not a critical error
	}

	// Create a new custom token for the user
	ctx := context.Background()
	customToken, err := firebaseClient.CustomToken(ctx, storedToken.UserID)
	if err != nil {
		fmt.Println("Failed to create custom token:", err)
		models.WriteServiceError(w, "Failed to generate new token", false, false, http.StatusInternalServerError)
		return
	}

	// Generate a new refresh token (rotate the refresh token for security)
	newRefreshTokenString, err := models.GenerateRefreshToken()
	if err != nil {
		fmt.Println("Failed to generate new refresh token:", err)
		models.WriteServiceError(w, "Failed to generate new refresh token", false, false, http.StatusInternalServerError)
		return
	}

	// Create a new refresh token with 7 days expiration
	newRefreshToken := &models.RefreshToken{
		UserID:    storedToken.UserID,
		Token:     newRefreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		UserAgent: r.Header.Get("User-Agent"),
		IPAddress: getClientIP(r),
	}

	// Save the new refresh token
	err = newRefreshToken.Save()
	if err != nil {
		fmt.Println("Failed to save new refresh token:", err)
		models.WriteServiceError(w, "Failed to save new refresh token", false, false, http.StatusInternalServerError)
		return
	}

	// Revoke the old refresh token
	err = storedToken.Revoke()
	if err != nil {
		fmt.Println("Failed to revoke old refresh token:", err)
		// Continue anyway - the new token is already created
	}

	// Return the new tokens
	response := RefreshTokenResponse{
		CustomToken:  customToken,
		RefreshToken: newRefreshTokenString,
		ExpiresIn:    3600, // Firebase custom tokens are valid for 1 hour
	}

	models.WriteServiceResponse(w, "Token refreshed successfully", response, true, true, http.StatusOK)
	fmt.Println("---Token refresh successful---")
}

// HandleRevokeToken handles POST requests to revoke a refresh token (logout)
func HandleRevokeToken(w http.ResponseWriter, r *http.Request) {
	fmt.Println("---HandleRevokeToken---")

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		models.WriteServiceError(w, "Failed to read request body", false, false, http.StatusBadRequest)
		return
	}

	var revokeReq RefreshTokenRequest
	err = json.Unmarshal(body, &revokeReq)
	if err != nil {
		models.WriteServiceError(w, "Invalid request format", false, false, http.StatusBadRequest)
		return
	}

	if revokeReq.RefreshToken == "" {
		models.WriteServiceError(w, "Refresh token is required", false, false, http.StatusBadRequest)
		return
	}

	// Find the refresh token in the database
	storedToken, err := models.FindRefreshTokenByToken(revokeReq.RefreshToken)
	if err != nil {
		fmt.Println("Refresh token not found:", err)
		// Return success even if token not found (idempotent operation)
		models.WriteServiceResponse(w, "Token revoked successfully", nil, true, true, http.StatusOK)
		return
	}

	// Revoke the token
	err = storedToken.Revoke()
	if err != nil {
		fmt.Println("Failed to revoke token:", err)
		models.WriteServiceError(w, "Failed to revoke token", false, false, http.StatusInternalServerError)
		return
	}

	models.WriteServiceResponse(w, "Token revoked successfully", nil, true, true, http.StatusOK)
	fmt.Println("---Token revocation successful---")
}

// HandleRevokeAllTokens handles POST requests to revoke all refresh tokens for a user
func HandleRevokeAllTokens(w http.ResponseWriter, r *http.Request) {
	fmt.Println("---HandleRevokeAllTokens---")

	// Get user ID from the authenticated context (requires authentication middleware)
	// This is a protected endpoint
	userID := r.Context().Value("user_id")
	if userID == nil || userID.(string) == "" {
		models.WriteServiceError(w, "Authentication required", false, false, http.StatusUnauthorized)
		return
	}

	// Revoke all tokens for the user
	err := models.RevokeAllUserTokens(userID.(string))
	if err != nil {
		fmt.Println("Failed to revoke all tokens:", err)
		models.WriteServiceError(w, "Failed to revoke all tokens", false, false, http.StatusInternalServerError)
		return
	}

	models.WriteServiceResponse(w, "All tokens revoked successfully", nil, true, true, http.StatusOK)
	fmt.Println("---All tokens revoked successfully---")
}
