package main

import (
	// "encoding/json"
	// "io"
	"log"
	"net/http"
)

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	_, err := s.validateOAuthToken(r)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	search := r.URL.Path[len("/v1/user/search/"):]
	// process input here
	log.Printf("Search: %s", search)

	switch r.Method {
	case http.MethodGet:
		s.handleElasticSearch(w, r, search)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleElasticSearch(w http.ResponseWriter, r *http.Request, search string) {
	log.Printf("GET request for search: %s", search)
	_, err := w.Write([]byte("GET request for search: " + search))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
