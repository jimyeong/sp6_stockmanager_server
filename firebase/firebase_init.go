package firebase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"google.golang.org/api/option"
)

// Initialize Firebase app once at startup
func InitFirebaseApp() (*auth.Client, error) {
	ctx := context.Background()
	credJSON := os.Getenv("FIREBASE_CREDENTIALS")
	if credJSON == "" {
		fmt.Println("---FIREBASE_CREDENTIALS is not set---")
		return nil, errors.New("FIREBASE_CREDENTIALS is not set")
	}
	credBytes := []byte(credJSON)

	fmt.Println("---credBytes---", credBytes)
	opt := option.WithCredentialsJSON(credBytes)

	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, err
	}

	client, err := app.Auth(ctx)
	if err != nil {
		return nil, err
	}
	// app, err := firebase.NewApp(context.Background(), nil)
	// if err != nil {
	// 	log.Fatalf("error initializing app: %v\n", err)
	// }
	// client, err := app.Auth(context.Background())
	// if err != nil {
	// 	log.Fatalf("error initializing client: %v\n", err)
	// }

	return client, nil
}

type contextKey string

const (
	userIDKey         contextKey = "user_id"
	userEmailKey      contextKey = "user_email"
	userNameKey       contextKey = "user_name"
	userPhoneKey      contextKey = "user_phone"
	userPhotoKey      contextKey = "user_photo_url"
	userVerifiedKey   contextKey = "user_email_verified"
	userAnonymousKey  contextKey = "user_is_anonymous"
	userProviderIDKey contextKey = "user_provider_id"
	tokenClaimsKey    contextKey = "token_claims"
)

// Add this near the top of the file, after the imports
type TokenClaims struct {
	UID           string    `json:"uid"`
	Email         string    `json:"email"`
	DisplayName   string    `json:"displayName"`
	EmailVerified bool      `json:"emailVerified"`
	IsAnonymous   bool      `json:"isAnonymous"`
	PhoneNumber   string    `json:"phoneNumber"`
	PhotoURL      string    `json:"photoURL"`
	ProviderId    string    `json:"providerId"`
	LoginAt       time.Time `json:"loginAt"`
	CreatedAt     time.Time `json:"createdAt"`
	Aud           string    `json:"aud"`
}

// WithUserContext creates a context with user information
func WithUserContext(ctx context.Context, claims TokenClaims) context.Context {
	ctx = context.WithValue(ctx, userIDKey, claims.UID)
	ctx = context.WithValue(ctx, userEmailKey, claims.Email)
	ctx = context.WithValue(ctx, userNameKey, claims.DisplayName)
	ctx = context.WithValue(ctx, userPhoneKey, claims.PhoneNumber)
	ctx = context.WithValue(ctx, userPhotoKey, claims.PhotoURL)
	ctx = context.WithValue(ctx, userVerifiedKey, claims.EmailVerified)
	ctx = context.WithValue(ctx, userAnonymousKey, claims.IsAnonymous)
	ctx = context.WithValue(ctx, userProviderIDKey, claims.ProviderId)
	ctx = context.WithValue(ctx, tokenClaimsKey, claims)
	return ctx
}

func GetUserFromContext(ctx context.Context) (string, string, string) {
	userID := ctx.Value(userIDKey).(string)
	userEmail := ctx.Value(userEmailKey).(string)
	userName := ctx.Value(userNameKey).(string)
	return userID, userEmail, userName
}

func GetTokenClaimsFromContext(ctx context.Context) TokenClaims {
	return ctx.Value(tokenClaimsKey).(TokenClaims)
}
func SaveTokenClaimsToContext(ctx context.Context, claims TokenClaims) context.Context {
	return context.WithValue(ctx, tokenClaimsKey, claims)
}
