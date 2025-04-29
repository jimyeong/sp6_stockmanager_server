package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/jimyeongjung/owlverload_api/api/handlers"
	"github.com/jimyeongjung/owlverload_api/models"
	"github.com/joho/godotenv"
	"github.com/rs/cors"
)

func main() {
	var err error
	if os.Getenv("ENV") == "development" {
		err = godotenv.Load(".env.development")
	} else {
		err = godotenv.Load(".env.production")
	}
	if err != nil {
		log.Fatal("Error loading .env file")
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

	fmt.Println("@main@2", "Registering routes")
	// Register routes
	r.HandleFunc("/api/v1/auth/signin", handlers.HandleSignIn).Methods("POST")

	// Start server
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", r))
}
