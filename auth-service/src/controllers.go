package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"os"

	"google.golang.org/api/idtoken"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		DB: db,
	}
}

type authRequestData struct {
	Token string `json:"token"`
}

func validateToken(token string) (string, string, error) {
	oauthClient := os.Getenv("OAUTH_CLIENT")
	if oauthClient == "" {
		return "", "", errors.New("OAUTH_CLIENT not set")
	}
	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, token, oauthClient)
	if err != nil {
		return "", "", err
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return "", "", errors.New("email not found in token")
	}

	name, ok := payload.Claims["name"].(string)
	if !ok {
		return "", "", errors.New("name not found in token")
	}

	return email, name, nil
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	printRequestDetails(r)
	// get token
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		log.Println("Error reading request body:", err)
		return
	}
	defer r.Body.Close()

	var data authRequestData
	err = json.Unmarshal(body, &data)
	if err != nil {
		http.Error(w, "Error parsing request body", http.StatusBadRequest)
		log.Println("Error parsing request body:", err)
		return
	}

	// validate token + extract email
	email, name, err := validateToken(data.Token)
	if err != nil {
		http.Error(w, "Error validating token", http.StatusInternalServerError)
		log.Println("Error validating token:", err)
		return
	}
	log.Println("Email:", email)
	log.Println("Name:", name)

	// check if user exists in db
	var isActive bool
	err = s.DB.QueryRow("SELECT is_active FROM users WHERE email = $1", email).Scan(&isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			// user does not exist, do create
			_, err = s.DB.Exec("INSERT INTO users (email, name, is_active) VALUES ($1, $2, $3)", email, name, true)
			if err != nil {
				http.Error(w, "Database error", http.StatusInternalServerError)
				log.Println("DB Insert Error:", err)
				return
			}
			w.WriteHeader(http.StatusCreated)
			_, err = w.Write([]byte(`{"active":false}`))
			if err != nil {
				log.Println("Error writing response:", err)
			}
			return
		}
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("DB Query Error:", err)
		return
	}

	// if exists and isCreated==true, return true
	w.WriteHeader(http.StatusOK)
	if isActive {
		_, err = w.Write([]byte(`{"active":true}`))
	} else {
		_, err = w.Write([]byte(`{"active":false}`))
	}
	if err != nil {
		log.Println("Error writing response:", err)
	}
}
