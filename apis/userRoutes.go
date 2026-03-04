package apis

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jimyeongjung/owlverload_api/models"
)

type UserRequest struct {
	User models.User `json:"user"`
}

// HandleSignIn handles POST requests to save user data
func HandleSignIn(w http.ResponseWriter, r *http.Request) {

	// Parse the request body
	body, err := io.ReadAll(r.Body)
	result := make([]models.User, 1)

	if err != nil {
		models.WriteServiceError(w, "Failed to save user: uid is empty", false, false, http.StatusBadRequest)
		if err != nil {
			log.Printf("Failed to write service error: %v", err)
		}
		return
	}

	var userRequest UserRequest
	err = json.Unmarshal(body, &userRequest)
	if err != nil {
		models.WriteServiceError(w, "Failed to save user: uid is empty", false, false, http.StatusBadRequest)
		if err != nil {
			log.Printf("Failed to write service error: %v", err)
		}
		return
	}

	user := userRequest.User
	fmt.Println("user params:", user)

	if user.Uid == "" {
		models.WriteServiceError(w, "Failed to save user: uid is empty", false, false, http.StatusBadRequest)
		if err != nil {
			log.Printf("Failed to write service error: %v", err)
		}
		return
	}
	// check if user exists
	isSaved := user.IsUserSaved(user.Uid)

	if isSaved {
		// update login time
		user.LoginAt = time.Now()
		user, err = user.Update(user.Uid)
		if err != nil {
			fmt.Println("err:", err)
		}
	} else {
		// save user
		user, err = user.Save()
		if err != nil {
			// http.Error(w, "Failed to save user", http.StatusInternalServerError)
			models.WriteServiceError(w, "Failed to save user", false, false, http.StatusInternalServerError)
			if err != nil {
				log.Printf("Failed to write service error: %v", err)
			}
			return
		}
	}

	result[0] = user

	// Generate refresh token for the user
	refreshTokenString, err := models.GenerateRefreshToken()
	if err != nil {
		fmt.Println("Failed to generate refresh token:", err)
		models.WriteServiceError(w, "Failed to generate refresh token", false, false, http.StatusInternalServerError)
		return
	}

	// Create refresh token with 7 days expiration
	refreshToken := &models.RefreshToken{
		UserID:    user.Uid,
		Token:     refreshTokenString,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		UserAgent: r.Header.Get("User-Agent"),
		IPAddress: getClientIP(r),
	}

	// Save the refresh token
	err = refreshToken.Save()
	if err != nil {
		fmt.Println("Failed to save refresh token:", err)
		models.WriteServiceError(w, "Failed to save refresh token", false, false, http.StatusInternalServerError)
		return
	}

	// Prepare response with user data and refresh token
	responseData := map[string]interface{}{
		"user":          user,
		"refresh_token": refreshTokenString,
		"expires_in":    7 * 24 * 60 * 60, // 7 days in seconds
	}

	// return user data with refresh token
	models.WriteServiceResponse(w, "Success", responseData, true, true, http.StatusOK)
	fmt.Println("---User response sent with refresh token---")
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		// X-Forwarded-For can contain multiple IPs, get the first one
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	ip := r.RemoteAddr
	// RemoteAddr includes port, strip it
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}

	return ip
}
