package main

import (
	"log"
	"net/http"

	"user-auth/internal/auth"
	"user-auth/internal/db"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

const (
	host     = "db"
	port     = 5432 // Держим port как int
	user     = "postgres"
	password = "swagimpoe123"
	dbname   = "auth_db"
)

func main() {
	// Инициализация базы данных
	database := db.InitDB(host, port, user, password, dbname) // Передаем port как int
	defer database.Close()                                    // Закрыть соединение после использования

	auth.SetDB(database) // Установить базу данных в пакет auth

	r := mux.NewRouter()

	r.HandleFunc("/register", auth.RegisterUser).Methods("POST")
	r.HandleFunc("/login", auth.LoginUser).Methods("POST")
	r.HandleFunc("/users", auth.TokenValid(auth.GetAllUsers)).Methods("GET")
	r.HandleFunc("/user/{id}", auth.TokenValid(auth.GetUserByID)).Methods("GET")
	r.HandleFunc("/refresh", auth.RefreshAccessToken).Methods("POST")

	allowedOrigins := handlers.AllowedOrigins([]string{"http://localhost:3000"})
	allowedMethods := handlers.AllowedMethods([]string{"POST", "GET", "OPTIONS"})
	allowedHeaders := handlers.AllowedHeaders([]string{"Content-Type", "Authorization"})

	log.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handlers.CORS(allowedOrigins, allowedMethods, allowedHeaders)(r)))
}
