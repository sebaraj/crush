package main

import (
	// "encoding/json"
	// "io"
	"log"
	"net/http"
)

func (s *Server) handleUser(w http.ResponseWriter, r *http.Request) {
	email := r.URL.Path[len("/v1/user/info/"):]
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	log.Printf("Email: %s", email)
	emailFromToken, err := s.validateOAuthToken(r)
	if err != nil || emailFromToken != email {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleGetUser(w, r, email)
	case http.MethodPut:
		s.handleUpdateUser(w, r, email)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("GET request for user: %s", email)
	_, err := w.Write([]byte("GET request for user: " + email))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("PUT request to update user: %s", email)
	_, err := w.Write([]byte("PUT request to update user: " + email))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
