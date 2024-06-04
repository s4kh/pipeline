package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

type Conn interface {
	Get() *sql.DB
	Close() error
}

type DB struct {
	*sql.DB
}

func NewPostgresConnection(url string) (Conn, error) {
	dbConn, err := sql.Open("postgres", url)

	if err != nil {
		fmt.Println(err)
		return nil, fmt.Errorf("error connecting to DB: %s, %v", url, err)
	}

	log.Println(dbConn.Stats().InUse)

	return &DB{DB: dbConn}, err
}

func (db *DB) Get() *sql.DB {
	return db.DB
}

func (db *DB) Close() error {
	if db.DB != nil {
		return db.DB.Close()
	}

	return fmt.Errorf("cannot close nil db")
}
