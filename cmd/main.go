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
	port, _ := strconv.Atoi(os.Getenv("DATABASE_PORT"))
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")

	// Логирование параметров подключения к базе данных
	log.Println("Database Connection Parameters:")
	log.Printf("Host: %s", host)
	log.Printf("Port: %d", port)
	log.Printf("User: %s", user)
	log.Printf("Password: %s", password) // Будьте осторожны с логированием паролей!
	log.Printf("Database Name: %s", dbname)

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
