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
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true", dbConfig.DB_USER, dbConfig.DB_PASSWORD, dbConfig.DB_HOST, dbConfig.DB_PORT, dbConfig.DB_NAME))
	if err != nil {
		log.Fatal(err)
		return &SQLDB{DB: nil, Err: err}
	}
	fmt.Println("---DB connected---")
	return &SQLDB{DB: db, Err: nil}
}

func (s *SQLDB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return s.DB.Query(query, args...)
}

func (s *SQLDB) QueryRow(query string, args ...interface{}) *sql.Row {
	return s.DB.QueryRow(query, args...)
}

func (s *SQLDB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.DB.Exec(query, args...)
}

func (s *SQLDB) Prepare(query string) (*sql.Stmt, error) {
	return s.DB.Prepare(query)
}

func GetDBConfig() DBConfig {
	return DBConfig{
		DB_USER:     os.Getenv("DB_USER"),
		DB_PASSWORD: os.Getenv("DB_PASSWORD"),
		DB_HOST:     os.Getenv("DB_HOST"),
		DB_PORT:     os.Getenv("DB_PORT"),
		DB_NAME:     os.Getenv("DB_NAME"),
	}
}

func GetDBInstance(dbConfig DBConfig) *sql.DB {
	once.Do(func() {
		dbInstance = NewSQLDB(dbConfig).DB
	})
	return dbInstance
}
