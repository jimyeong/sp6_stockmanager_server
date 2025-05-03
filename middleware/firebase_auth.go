package middleware

import (
	"context"
	"fmt"
	"strings"
)

// FirebaseTokenValidator validates Firebase authentication tokens
// In a real implementation, you would use Firebase Admin SDK to verify tokens
// This is a placeholder that can be replaced with actual Firebase token validation
func FirebaseTokenValidator(token string) (bool, string, error) {
	// IMPORTANT: This is only a placeholder!
	// In a real implementation, you would use Firebase Admin SDK like this:
	//
	// import (
	//     "firebase.google.com/go/auth"
	// )
	//
	// func VerifyFirebaseToken(token string) (bool, string, error) {
	//     ctx := context.Background()
	//     client, err := firebaseApp.Auth(ctx)
	//     if err != nil {
	//         return false, "", err
	//     }
	//
	//     decodedToken, err := client.VerifyIDToken(ctx, token)
	//     if err != nil {
	//         return false, "", err
	//     }
	//
	//     return true, decodedToken.UID, nil
	// }

	// For testing purposes, let's implement a very simple validator
	// Remove this in production and implement proper validation
	if token == "" {
		return false, "", fmt.Errorf("token is empty")
	}

	// Check if token has a reasonable format
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return false, "", fmt.Errorf("invalid token format")
	}

	// In a real implementation, you would verify the token signature
	// and decode the payload to get the user ID
	// For now, just return a successful validation
	return true, "firebase-user-id", nil
}

// GetAuthToken extracts the token from Authorization header
func GetAuthToken(authHeader string) (string, error) {
	if authHeader == "" {
		return "", fmt.Errorf("missing Authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", fmt.Errorf("invalid Authorization format")
	}

	token := parts[1]
	if token == "" {
		return "", fmt.Errorf("empty token")
	}

	return token, nil
}

// WithUserContext creates a context with user information
func WithUserContext(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "userID", userID)
}