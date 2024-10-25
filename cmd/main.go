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
	// Get database host from environment variable; default to "localhost"
	host := os.Getenv("DATABASE_HOST")
	if host == "" {
		host = "localhost" // Default for local development
	}

	// Get database port from environment variable
	portStr := os.Getenv("DATABASE_PORT")
	port := 5432 // Default PostgreSQL port
	if portStr != "" {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			log.Fatalf("Invalid port number: %v", err)
		}
	}

	// Get database user from environment variable; default to "postgres"
	user := os.Getenv("DATABASE_USER")
	if user == "" {
		user = "postgres"
	}

	// Get database password from environment variable; default to a hardcoded password
	password := os.Getenv("DATABASE_PASSWORD")
	if password == "" {
		password = "swagimpoe123" // Be cautious with hardcoded passwords
	}

	// Get database name from environment variable; default to "auth_db"
	dbname := os.Getenv("DATABASE_NAME")
	if dbname == "" {
		dbname = "auth_db"
	}

	// Initialize database connection
	database := db.InitDB(host, port, user, password, dbname)
	defer database.Close()

	// Set the database instance in the auth package
	auth.SetDB(database)

	// Create a new router
	r := mux.NewRouter()

	// Define routes and their handlers
	r.HandleFunc("/register", auth.RegisterUser).Methods("POST")
	r.HandleFunc("/login", auth.LoginUser).Methods("POST")
	r.HandleFunc("/users", auth.TokenValid(auth.GetAllUsers)).Methods("GET")
	r.HandleFunc("/user/{id}", auth.TokenValid(auth.GetUserByID)).Methods("GET")
	r.HandleFunc("/refresh", auth.RefreshAccessToken).Methods("POST")

	// Configure CORS
	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	allowedMethods := handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	// Start the server
	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)))
}
