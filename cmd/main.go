package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"user-auth/internal/auth"
	"user-auth/internal/db"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		host = "localhost"
	}

	portStr := os.Getenv("DATABASE_PORT")
	port := 5432
	if portStr != "" {
		port, _ = strconv.Atoi(portStr)
	}

	user := os.Getenv("DATABASE_USER")
	if user == "" {
		user = "postgres"
	}

	password := os.Getenv("DATABASE_PASSWORD")
	if password == "" {
		password = "swagimpoe123"
	}

	dbname := os.Getenv("DATABASE_NAME")
	if dbname == "" {
		dbname = "auth_db"
	}

	database := db.InitDB(host, port, user, password, dbname)
	defer database.Close()

	db.CreateTables(database)

	auth.SetDB(database)

	r := mux.NewRouter()

	r.HandleFunc("/register", auth.RegisterUser).Methods("POST")
	r.HandleFunc("/login", auth.LoginUser).Methods("POST")
	r.HandleFunc("/users", auth.TokenValid(auth.GetAllUsers)).Methods("GET")
	r.HandleFunc("/user/{id}", auth.TokenValid(auth.GetUserByID)).Methods("GET")
	r.HandleFunc("/refresh", auth.RefreshAccessToken).Methods("POST")

	allowedOrigins := handlers.AllowedOrigins([]string{"http://nbhtn.s3-website-us-east-1.amazonaws.com"})
	allowedMethods := handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)))
}
