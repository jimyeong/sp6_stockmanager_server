package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jimyeongjung/owlverload_api/apis"
	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/middleware"
	"github.com/jimyeongjung/owlverload_api/models"
	"github.com/jimyeongjung/owlverload_api/utils"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	var err error

	// Initialize logger
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("Failed to create log directory: %v", err)
	}

	logFile := filepath.Join(logDir, "owlverload_api.log")
	if err := utils.InitLogger(logFile); err != nil {
		log.Printf("Failed to initialize logger: %v", err)
	}
	defer utils.Close()

	// Set log level (Debug to see all logs)
	utils.SetLogLevel(utils.LevelDebug)

	utils.Info("Starting Owlverload API server")

	// Load environment variables
	if os.Getenv("ENV") == "development" {
		err = godotenv.Load(".env.development")
		utils.Info("Loaded development environment")
	} else {
		err = godotenv.Load(".env.production")
		utils.Info("Loaded production environment")
	}
	if err != nil {
		utils.Fatal("Error loading .env file: %v", err)
	}

	// db connection
	DB_USER := os.Getenv("DB_USER")
	DB_PASSWORD := os.Getenv("DB_PASSWORD")
	DB_HOST := os.Getenv("DB_HOST")
	DB_PORT := os.Getenv("DB_PORT")
	DB_NAME := os.Getenv("DB_NAME")
	dbConfig := models.DBConfig{
		DB_USER:     DB_USER,
		DB_PASSWORD: DB_PASSWORD,
		DB_HOST:     DB_HOST,
		DB_PORT:     DB_PORT,
		DB_NAME:     DB_NAME,
	}
	db := models.NewSQLDB(dbConfig)
	if db.Err != nil {
		log.Fatal(db.Err)
	}

	// router
	r := mux.NewRouter()

	// cors
	r.Use(cors.AllowAll().Handler)

	// Initialize Firebase app
	firebaseClient, err := firebase.InitFirebaseApp()
	if err != nil {
		log.Fatal(err)
	}

	// Define authentication middleware
	// authConfig := middleware.AuthenticationConfig{
	// 	ValidateToken:     middleware.FirebaseTokenValidator,
	// 	ExcludedPaths:     []string{"/public/api/v1/auth/signin", "/public/health", "/public/api/v1/health"},
	// 	TokenErrorMessage: "Authentication required. Please provide a valid Bearer token.",
	// }

	// Apply authentication middleware to protected routes
	// You can use either the original middleware or the new ValidateFirebaseToken middleware

	// Option 1: Original authentication middleware
	// apiRouter.Use(middleware.NewAuthentication(authConfig))

	// Option 2: New token validation middleware (simpler and more focused)

	// Create a subrouter for protected routes
	apiRouter := r.PathPrefix("/api/v1/").Subrouter()
	apiRouter.Use(func(next http.Handler) http.Handler {
		fmt.Println("--- coming in here 2--- ")
		return middleware.ValidateFirebaseToken(next, firebaseClient)
	})

	fmt.Println("@main@2", "Registering routes")

	// Public routes (no authentication required)
	r.HandleFunc("/public/api/v1/auth/signin", apis.HandleSignIn).Methods("POST")
	r.HandleFunc("/public/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	r.HandleFunc("/public/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Protected routes (authentication required)
	// Stock/Item routes
	fmt.Println("--- coming in here --- ")
	// apiRouter.HandleFunc("/getItem", apis.HandleGetItemByBarcode).Methods("GET")
	apiRouter.HandleFunc("/createNewItem", apis.HandleCreateItem).Methods("POST")
	// apiRouter.HandleFunc("/editItem/{itemId}", apis.HandleGetItemById).Methods("GET")
	apiRouter.HandleFunc("/getItemByBarcode", apis.HandleGetItemByBarcode).Methods("GET")
	apiRouter.HandleFunc("/getItemByCode", apis.HandleGetItemByCode).Methods("GET")
	apiRouter.HandleFunc("/getItemById", apis.HandleGetItemById).Methods("GET")
	apiRouter.HandleFunc("/updateItemById", apis.HandleUpdateItemById).Methods("PUT")
	apiRouter.HandleFunc("/stockIn", apis.HandleStockIn).Methods("POST")
	apiRouter.HandleFunc("/stockOut", apis.HandleStockOut).Methods("POST")
	apiRouter.HandleFunc("/stockUpdate", apis.HandleStockUpdate).Methods("PUT")
	// apiRouter.HandleFunc("/createItem", apis.HandleCreateItem).Methods("POST")
	apiRouter.HandleFunc("/registerItem", apis.HandleRegisterItem).Methods("POST")
	apiRouter.HandleFunc("/updateItem", apis.HandleUpdateItem).Methods("PUT")
	apiRouter.HandleFunc("/getItems", apis.HandleGetItems).Methods("GET")
	apiRouter.HandleFunc("/getItemsPaginated", apis.HandleGetItemsPaginated).Methods("GET")
	apiRouter.HandleFunc("/searchItems", apis.HandleSearchItems).Methods("POST")
	apiRouter.HandleFunc("/getItemsWithMissingInfo", apis.HandleGetItemsWithMissingInfo).Methods("GET")
	apiRouter.HandleFunc("/lookupItems", apis.HandleLookupItems).Methods("POST")
	apiRouter.HandleFunc("/getItemsExpiringWithinDays", apis.HandleGetItemsExpiringWithinDays).Methods("GET")

	// Tag routes
	apiRouter.HandleFunc("/tags", apis.HandleGetAllTags).Methods("GET")
	apiRouter.HandleFunc("/tags/create", apis.HandleCreateTag).Methods("POST")
	apiRouter.HandleFunc("/tags/popular", apis.HandleGetPopularTags).Methods("GET")
	apiRouter.HandleFunc("/tags/search", apis.HandleSearchTags).Methods("GET")
	apiRouter.HandleFunc("/tags/item/{itemId}", apis.HandleGetTagsForItem).Methods("GET")
	apiRouter.HandleFunc("/tags/associate", apis.HandleAssociateItemWithTags).Methods("POST")
	apiRouter.HandleFunc("/recommendations", apis.HandleGetRecommendedItems).Methods("POST")

	// Barcode routes
	apiRouter.HandleFunc("/saveBarcode", apis.HandleSaveBarcode).Methods("POST")

	// AI Helper routes
	apiRouter.HandleFunc("/analyze_barcode", apis.HandleBarcodeAnalyze).Methods("POST")

	// Image upload routes
	apiRouter.HandleFunc("/upload/image", apis.HandleImageUpload).Methods("POST")
	apiRouter.HandleFunc("/delete/image", apis.HandleImageDelete).Methods("DELETE")

	// Start server
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
