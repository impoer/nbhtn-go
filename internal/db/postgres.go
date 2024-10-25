package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func InitDB(host string, port int, user, password, dbname string) *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to the database: %v", err)
	}

	log.Println("Successfully connected to the database!")
	return db
}

func CreateTables(db *sql.DB) {
	createUsersTable := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
        name VARCHAR(100) NOT NULL,
        email VARCHAR(100) UNIQUE NOT NULL,
        password VARCHAR(100) NOT NULL
    );`

	_, err := db.Exec(createUsersTable)
	if err != nil {
		log.Fatalf("Error creating users table: %v", err)
	}

	log.Println("Users table created successfully!")
}
