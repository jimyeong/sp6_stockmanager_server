package models

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
)

var (
	dbInstance *sql.DB
	once       sync.Once
)

type DBConfig struct {
	DB_USER     string
	DB_PASSWORD string
	DB_HOST     string
	DB_PORT     string
	DB_NAME     string
}

type SQLDB struct {
	DB  *sql.DB
	Err error
}

func NewSQLDB(dbConfig DBConfig) *SQLDB {
	// Trace function entry and exit for debugging
	fmt.Println("@@@@@@@@@@@@@@@@@@---NewSQLDB---", dbConfig)
	// Log connection string without password for security
	connStrRedacted := fmt.Sprintf("%s:***@tcp(%s:%s)/%s?parseTime=true",
		dbConfig.DB_USER, dbConfig.DB_HOST, dbConfig.DB_PORT, dbConfig.DB_NAME)
	fmt.Println("@@@@@@@@@@@@@@@@@@---connStrRedacted---", connStrRedacted)
	// Actual connection string with password
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
		dbConfig.DB_USER, dbConfig.DB_PASSWORD, dbConfig.DB_HOST, dbConfig.DB_PORT, dbConfig.DB_NAME)
	fmt.Println("@@@@@@@@@@@@@@@@@@---connStr---", connStr)
	fmt.Println("@@@@@@@@@@@@@@@@@@---dbConfig.DB_HOST---", dbConfig.DB_HOST)
	fmt.Println("@@@@@@@@@@@@@@@@@@---dbConfig.DB_PORT---", dbConfig.DB_PORT)
	fmt.Println("@@@@@@@@@@@@@@@@@@---dbConfig.DB_NAME---", dbConfig.DB_NAME)
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		fmt.Println("@@@@@@@@@@@@@@@@@@---err---", err)
		log.Fatal(err)
		return &SQLDB{DB: nil, Err: err}
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		fmt.Println("@@@@@@@@@@@@@@@@@@---err---", err)
		return &SQLDB{DB: nil, Err: err}
	}
	fmt.Println("@@@@@@@@@@@@@@@@@@---db---", db)
	return &SQLDB{DB: db, Err: nil}
}

// Enhanced Query method with logging
func (s *SQLDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	fmt.Println("Query: %s\nArgs: %v", query, args)
	rows, err := s.DB.Query(query, args...)
	if err != nil {
		fmt.Println("Query failed: %v\nQuery: %s\nArgs: %v", err, query, args)
	}
	return rows, err
}

// Enhanced QueryRow method with logging
func (s *SQLDB) QueryRow(query string, args ...interface{}) *sql.Row {
	row := s.DB.QueryRow(query, args...)
	if row != nil {
		fmt.Println("QueryRow failed: %v\nQuery: %s\nArgs: %v", row, query, args)
	}
	return row
}

// Enhanced Exec method with logging
func (s *SQLDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	fmt.Println("Query: %s\nArgs: %v", query, args)
	result, err := s.DB.Exec(query, args...)
	if err != nil {
		fmt.Println("Exec failed: %v\nQuery: %s\nArgs: %v", err, query, args)
	} else {
		// Log affected rows for INSERT/UPDATE/DELETE queries
		if rowsAffected, err := result.RowsAffected(); err == nil {
			fmt.Println("Query affected %d rows", rowsAffected)
		} else {
			fmt.Println("Query affected %d rows", rowsAffected)
		}
	}
	return result, err
}

// Enhanced Prepare method with logging
func (s *SQLDB) Prepare(query string) (*sql.Stmt, error) {
	fmt.Println("Prepare: %s", query)
	stmt, err := s.DB.Prepare(query)
	if err != nil {
		fmt.Println("Prepare failed: %v\nQuery: %s", err, query)
	}
	return stmt, err
}

func GetDBConfig() DBConfig {
	fmt.Println("GetDBConfig")
	config := DBConfig{
		DB_USER:     os.Getenv("DB_USER"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_PORT:     os.Getenv("DB_PORT"),
		DB_NAME:     os.Getenv("DB_NAME"),
	}

	// Verify that we have all required configuration
	if config.DB_USER == "" || config.DB_HOST == "" || config.DB_PORT == "" || config.DB_NAME == "" {
		fmt.Println("Incomplete database configuration: User=%s, Host=%s, Port=%s, Name=%s",
			maskEmpty(config.DB_USER), maskEmpty(config.DB_HOST),
			maskEmpty(config.DB_PORT), maskEmpty(config.DB_NAME))
	}

	return config
}

// maskEmpty returns "<empty>" if the string is empty, otherwise returns the string
func maskEmpty(s string) string {
	if s == "" {
		return "<empty>"
	}
	return s
}

func GetDBInstance(dbConfig DBConfig) *sql.DB {
	fmt.Println("GetDBInstance")
	once.Do(func() {
		fmt.Println("NewSQLDB")
		dbInstance = NewSQLDB(dbConfig).DB
	})

	if dbInstance == nil {
		fmt.Println("Database instance is nil - initialization may have failed")
		return nil
	}

	return dbInstance
}
