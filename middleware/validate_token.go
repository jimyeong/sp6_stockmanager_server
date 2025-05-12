package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"firebase.google.com/go/auth"
	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/models"
)

// ValidateFirebaseToken is a middleware function that checks if a request has a valid Firebase token
// It will send a 401 Unauthorized response if the token is missing, expired, or invalid
func ValidateFirebaseToken(next http.Handler, client *auth.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		ctx := context.Background()
		// Extract Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			models.WriteServiceError(w, "Authentication required. Please provide a valid Bearer token.", false, false, http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			models.WriteServiceError(w, "Invalid authorization format. Use 'Bearer [token]'", false, false, http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			models.WriteServiceError(w, "Empty token provided", false, false, http.StatusUnauthorized)
			return
		}

		// Validate the token
		claims, err := client.VerifyIDToken(ctx, token)
		if err != nil {
			// Provide a specific error message based on the error
			errorMessage := "Invalid or expired token"

			if err != nil {
				switch {
				case strings.Contains(err.Error(), "expired"):
					errorMessage = "Token has expired. Please sign in again."
				case strings.Contains(err.Error(), "invalid signature"):
					errorMessage = "Invalid token signature"
				case strings.Contains(err.Error(), "invalid token"):
					errorMessage = "Invalid token format"
				default:
					errorMessage = fmt.Sprintf("Authentication error: %v", err)
				}
			}
			models.WriteServiceError(w, errorMessage, false, false, http.StatusUnauthorized)
			return
		}

		// If we get here, the token is valid
		// Create user object from claims
		user := models.User{
			Uid:           claims.UID,
			Email:         claims.Claims["email"].(string),
			DisplayName:   claims.Claims["name"].(string),
			EmailVerified: claims.Claims["email_verified"].(bool),
			PhotoURL:      claims.Claims["picture"].(string),
		}
		// Check if user already exists in the database and create/update as needed

		if user.IsUserSaved(user.Uid) {
			// Update existing user's login time
			_, err = user.Update(user.Uid)
			if err != nil {
				fmt.Println("Error updating user:", err)
				// Continue anyway - don't fail the request just because we couldn't update login time
			}
		} else {
			// Save new user
			_, err = user.Save()
			if err != nil {
				fmt.Println("Error saving new user:", err)
				// Continue anyway - don't fail the request just because we couldn't save the user
			}
		}

		// Add token information to request context
		tokenClaims := firebase.TokenClaims{
			UID:           claims.UID,
			Email:         claims.Claims["email"].(string),
			DisplayName:   claims.Claims["name"].(string),
			EmailVerified: claims.Claims["email_verified"].(bool),
			PhotoURL:      claims.Claims["picture"].(string),
		}
		ctx = firebase.WithUserContext(r.Context(), tokenClaims)

		// Pass the authenticated request to the next handler
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UseFirebaseAuth wraps a handler with Firebase authentication
// It's a convenience function to apply the ValidateFirebaseToken middleware
func UseFirebaseAuth(handler http.HandlerFunc, client *auth.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ValidateFirebaseToken(http.HandlerFunc(handler), client).ServeHTTP(w, r)
	}
}

// Simplified example of how Firebase token verification would work
// This is just for reference - in production, you'd need to initialize Firebase Admin SDK
/*
To implement this with the real Firebase Admin SDK, you would:
1. Add Firebase Admin SDK to your dependencies
2. Initialize the Firebase app in your main.go
3. Use the Auth client to verify tokens

Example implementation with Firebase Admin SDK:

*/

// Verify token using Firebase Admin SDK
func VerifyFirebaseIDToken(ctx context.Context, token string, client *auth.Client) (firebase.TokenClaims, error) {
	decodedToken, err := client.VerifyIDToken(ctx, token)
	if err != nil {
		return firebase.TokenClaims{}, err
	}

	// Extract claims from the token
	claims := firebase.TokenClaims{
		UID:           decodedToken.UID,
		Email:         decodedToken.Claims["email"].(string),
		DisplayName:   decodedToken.Claims["name"].(string),
		EmailVerified: decodedToken.Claims["email_verified"].(bool),
		IsAnonymous:   decodedToken.Claims["is_anonymous"].(bool),
		PhoneNumber:   decodedToken.Claims["phone_number"].(string),
		PhotoURL:      decodedToken.Claims["photo_url"].(string),
		ProviderId:    decodedToken.Claims["provider_id"].(string),
		LoginAt:       time.Now(),
		CreatedAt:     time.Now(),
		// Map other fields from decodedToken.Claims
	}

	return claims, nil
}
