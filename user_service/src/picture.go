package main

import (
	// "encoding/json"
	// "io"
	"log"
	"net/http"
)

func (s *Server) handlePicture(w http.ResponseWriter, r *http.Request) {
	printRequestDetails(r)
	email := r.URL.Path[len("/v1/user/picture/"):]
	log.Printf("Search: %s", email)

	emailFromToken, err := s.validateOAuthToken(r)
	if err != nil || email != emailFromToken {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleS3URLRequest(w, r, email)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleS3URLRequest(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("GET request for S3 signed url: %s", email)
	_, err := w.Write([]byte("GET request for S3 signed url: " + email))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
