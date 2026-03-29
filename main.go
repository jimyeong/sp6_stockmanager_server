package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jimyeongjung/owlverload_api/apis"
	"github.com/jimyeongjung/owlverload_api/firebase"
	"github.com/jimyeongjung/owlverload_api/middleware"
	"github.com/jimyeongjung/owlverload_api/models"
	v1Controller "github.com/jimyeongjung/owlverload_api/v1/controller"

	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	var err error

	// Initialize logger
	// logDir := "./logs"
	// if err := os.MkdirAll(logDir, 0755); err != nil {
	// 	log.Printf("Failed to create log directory: %v", err)
	// }

	// logFile := filepath.Join(logDir, "owlverload_api.log")
	// if err := utils.InitLogger(logFile); err != nil {
	// 	log.Printf("Failed to initialize logger: %v", err)
	// }
	// defer utils.Close()

	// Set log level (Debug to see all logs)
	// utils.SetLogLevel(utils.LevelDebug)

	// Load environment variables
	if os.Getenv("ENV") == "development" {
		err = godotenv.Load(".env.development")
		fmt.Println("---Loading .env.development---")
	} else if os.Getenv("ENV") == "staging" {
		err = godotenv.Load(".env.staging")
		fmt.Println("---Loading .env.staging---")
	} else {
		err = godotenv.Load(".env.production")
		fmt.Println("---Loading .env.production---")
	}
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
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
	// Create a subrouter for protected routes
	apiRouter := r.PathPrefix("/api/v1/").Subrouter()
	apiRouter.Use(func(next http.Handler) http.Handler {
		return middleware.ValidateFirebaseToken(next, firebaseClient)
	})
	// r.Use(middleware.IdempotencyMiddleware)

	// Public routes (no authentication required)
	r.HandleFunc("/public/api/v1/auth/signin", apis.HandleSignIn).Methods("POST")
	r.HandleFunc("/public/api/v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		apis.HandleRefreshToken(w, r, firebaseClient)
	}).Methods("POST")
	r.HandleFunc("/public/api/v1/auth/revoke", apis.HandleRevokeToken).Methods("POST")
	r.HandleFunc("/public/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")
	r.HandleFunc("/public/api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// apiRouter.HandleFunc("/createNewItem", apis.HandleCreateItem).Methods("POST")
	//apiRouter.HandleFunc("/getItemByBarcode", apis.HandleGetItemByBarcode).Methods("GET")
	//apiRouter.HandleFunc("/getItemByCode", apis.HandleGetItemByCode).Methods("GET")
	//apiRouter.HandleFunc("/getItemById", apis.HandleGetItemById).Methods("GET")
	//apiRouter.HandleFunc("/updateItemById", apis.HandleUpdateItemById).Methods("PUT")

	//apiRouter.HandleFunc("/stockOut", apis.HandleStockOut).Methods("POST")
	//apiRouter.HandleFunc("/stockUpdate", apis.HandleStockUpdate).Methods("PUT")
	// apiRouter.HandleFunc("/stock/{stockId}", apis.HandleStockDeleteById).Methods("DELETE")
	//apiRouter.HandleFunc("/registerItem", apis.HandleRegisterItem).Methods("POST")
	//apiRouter.HandleFunc("/updateItem", apis.HandleUpdateItem).Methods("PUT")
	// apiRouter.HandleFunc("/getItems", apis.HandleGetItems).Methods("GET")
	//apiRouter.HandleFunc("/getItemsPaginated", apis.HandleGetItemsPaginated).Methods("GET")
	//apiRouter.HandleFunc("/searchItems", apis.HandleSearchItems).Methods("POST")
	//apiRouter.HandleFunc("/getItemsWithMissingInfo", apis.HandleGetItemsWithMissingInfo).Methods("GET")

	// apiRouter.HandleFunc("/getItemsExpiringWithinDays", apis.HandleGetItemsExpiringWithinDays).Methods("GET")
	// apiRouter.HandleFunc("/getItemsWithExpiredStocksOlderThanDays", apis.HandleGetItemsWithExpiredStocksOlderThanDays).Methods("GET")
	// apiRouter.HandleFunc("/getItemsWithExpiredStocksAheadOfDays", apis.HandleGetItemsWithExpiredStocksAheadOfDays).Methods("GET")

	// just stock request
	// apiRouter.HandleFunc("/getProductStockByItemId", apis.HandleGetStockByItemId).Methods("GET")

	// apiRouter.HandleFunc("/tags", apis.HandleGetAllTags).Methods("GET")
	// apiRouter.HandleFunc("/tags/create", apis.HandleCreateTag).Methods("POST")
	// apiRouter.HandleFunc("/tags/popular", apis.HandleGetPopularTags).Methods("GET")
	// apiRouter.HandleFunc("/tags/search", apis.HandleSearchTags).Methods("GET")

	// apiRouter.HandleFunc("/tags/item/{itemId}", apis.HandleGetTagsForItem).Methods("GET")
	// apiRouter.HandleFunc("/tags/associate", apis.HandleAssociateItemWithTags).Methods("POST")
	// apiRouter.HandleFunc("/recommendations", apis.HandleGetRecommendedItems).Methods("POST")

	// Barcode routes
	// apiRouter.HandleFunc("/saveBarcode", apis.HandleSaveBarcode).Methods("POST")

	// AI Helper routes
	// apiRouter.HandleFunc("/analyze_barcode", apis.HandleBarcodeAnalyze).Methods("POST")

	// Image upload routes
	// apiRouter.HandleFunc("/upload/image", apis.HandleImageUpload).Methods("POST")
	apiRouter.HandleFunc("/delete/image", apis.HandleImageDelete).Methods("DELETE")

	// v1
	apiRouter.HandleFunc("/auth/revoke-all", apis.HandleRevokeAllTokens).Methods("POST")

	apiRouter.HandleFunc("/product/update", v1Controller.HandleProductUpdate).Methods("PUT")
	apiRouter.HandleFunc("/products/inventory/{productId}", v1Controller.HandleGetInventoryByProductId).Methods("GET")
	apiRouter.HandleFunc("/products/expiring-stocks", v1Controller.HandleGetProductsWithExpiringStocksByDateRange).Methods("GET")
	apiRouter.HandleFunc("/products/expiring-stocks-with-days-left", v1Controller.HandleGetProductsWithStockWithDaysLeft).Methods("GET")
	apiRouter.HandleFunc("/products/expired-inventory", v1Controller.HandleGetExpiredInventoryOlderThanDays).Methods("GET")
	apiRouter.HandleFunc("/products/finalise-expired-stock", v1Controller.HandleFinaliseExpiredStock).Methods("POST")
	apiRouter.HandleFunc("/inventory/search", v1Controller.HandleSearchInventory).Methods("POST")
	apiRouter.HandleFunc("/stocks/{productId}", v1Controller.HandleGetStocksByProductId).Methods("GET")
	apiRouter.HandleFunc("/stocks/create", v1Controller.HandleCreateStock).Methods("POST")

	apiRouter.HandleFunc("/image/upload", v1Controller.HandleImageUpload).Methods("POST")
	apiRouter.HandleFunc("/image/delete", v1Controller.HandleImageDelete).Methods("DELETE")

	// Start server
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
