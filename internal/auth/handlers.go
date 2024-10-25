package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

func generateJWT(email string) (string, error) {
	log.Printf("LoG")
	claims := jwt.MapClaims{}
	claims["authorized"] = true
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("mySecretKey"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoG1")
	w.Header().Set("Content-Type", "application/json")

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, jsonError("400", "Invalid JSON format"), http.StatusBadRequest)
		return
	}

	if user.Name == "" || user.Email == "" || user.Password == "" {
		http.Error(w, jsonError("400", "Missing required fields: Name, Email, and Password must not be empty"), http.StatusBadRequest)
		return
	}

	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)`, user.Email).Scan(&exists)
	if err != nil {
		http.Error(w, jsonError("500", "Database error: unable to check user existence"), http.StatusInternalServerError)
		return
	}

	if exists {
		http.Error(w, jsonError("409", "User already exists with this email"), http.StatusConflict)
		return
	}

	sqlStatement := `INSERT INTO users (name, email, password) VALUES ($1, $2, $3) RETURNING id`
	id := 0
	err = db.QueryRow(sqlStatement, user.Name, user.Email, user.Password).Scan(&id)
	if err != nil {
		http.Error(w, jsonError("500", "Database error: unable to insert user"), http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"message": "User registered successfully!",
		"user_id": id,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoG2")
	w.Header().Set("Content-Type", "application/json")

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, jsonError("400", "Invalid request: unable to decode JSON"), http.StatusBadRequest)
		return
	}

	var storedPassword string
	sqlStatement := `SELECT password FROM users WHERE email=$1`
	err := db.QueryRow(sqlStatement, user.Email).Scan(&storedPassword)
	if err == sql.ErrNoRows {
		http.Error(w, jsonError("401", "User not found: Invalid email or user does not exist"), http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, jsonError("500", "Database error: unable to retrieve user data"), http.StatusInternalServerError)
		return
	}

	if storedPassword != user.Password {
		http.Error(w, jsonError("401", "Invalid password: please check your credentials"), http.StatusUnauthorized)
		return
	}

	accessToken, err := generateJWT(user.Email)
	if err != nil {
		http.Error(w, jsonError("500", "Token generation failed: unable to create access token"), http.StatusInternalServerError)
		return
	}

	refreshToken, err := generateRefreshToken(user.Email)
	if err != nil {
		http.Error(w, jsonError("500", "Token generation failed: unable to create refresh token"), http.StatusInternalServerError)
		return
	}

	response := Tokens{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func jsonError(code string, message string) string {
	errorResponse := map[string]interface{}{
		"error":   code,
		"message": message,
	}
	jsonResponse, _ := json.Marshal(errorResponse)
	return string(jsonResponse)
}

func GetAllUsers(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoG3")
	w.Header().Set("Content-Type", "application/json")

	rows, err := db.Query(`SELECT id, name, email FROM users`)
	if err != nil {
		http.Error(w, jsonError("500", "Database error: unable to retrieve users"), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Name, &user.Email); err != nil {
			http.Error(w, jsonError("500", "Database error: unable to parse user data"), http.StatusInternalServerError)
			return
		}
		users = append(users, user)
	}

	if err = rows.Err(); err != nil {
		http.Error(w, jsonError("500", "Database error: row processing error"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func GetUserByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoG4")
	w.Header().Set("Content-Type", "application/json")

	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		http.Error(w, jsonError("400", "Missing user ID"), http.StatusBadRequest)
		return
	}

	var user User
	sqlStatement := `SELECT id, name, email FROM users WHERE id=$1`
	err := db.QueryRow(sqlStatement, id).Scan(&user.ID, &user.Name, &user.Email)
	if err == sql.ErrNoRows {
		http.Error(w, jsonError("404", "User not found"), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, jsonError("500", "Database error: unable to retrieve user"), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func generateRefreshToken(email string) (string, error) {
	log.Printf("LoG5")
	claims := jwt.MapClaims{}
	claims["email"] = email
	claims["exp"] = time.Now().Add(time.Hour * 24 * 7).Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("mySecretKey"))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func TokenValid(next http.HandlerFunc) http.HandlerFunc {
	log.Printf("LoG6")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, jsonError("401", "Missing authorization token"), http.StatusUnauthorized)
			return
		}

		const bearerPrefix = "Bearer "
		if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
			http.Error(w, jsonError("401", "Invalid authorization format"), http.StatusUnauthorized)
			return
		}
		tokenString := authHeader[len(bearerPrefix):]

		claims := &jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("mySecretKey"), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, jsonError("401", "Invalid token"), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "email", (*claims)["email"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RefreshAccessToken(w http.ResponseWriter, r *http.Request) {
	log.Printf("LoG7")
	w.Header().Set("Content-Type", "application/json")

	var tokens Tokens
	if err := json.NewDecoder(r.Body).Decode(&tokens); err != nil {
		http.Error(w, jsonError("400", "Invalid request: unable to decode JSON"), http.StatusBadRequest)
		return
	}

	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokens.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("mySecretKey"), nil
	})
	if err != nil {
		http.Error(w, jsonError("401", "Invalid refresh token"), http.StatusUnauthorized)
		return
	}

	email := (*claims)["email"].(string)
	accessToken, err := generateJWT(email)
	if err != nil {
		http.Error(w, jsonError("500", "Token generation failed: unable to create access token"), http.StatusInternalServerError)
		return
	}

	response := Tokens{
		AccessToken: accessToken,
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
