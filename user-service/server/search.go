package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
)

func (s *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
	printRequestDetails(r)
	_, err := s.validateOAuthToken(r)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// search := r.URL.Path[len("/v1/user/search/"):]

	switch r.Method {
	case http.MethodGet:
		s.handleElasticSearch(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleElasticSearch(w http.ResponseWriter, r *http.Request) {
	// copy json body from reader (which will be in proper opensearch format to send to opensearch)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if !json.Valid(body) {
		http.Error(w, "Invalid JSON in request body", http.StatusBadRequest)
		return
	}

	res, err := s.OpenSearchClient.Search(
		s.OpenSearchClient.Search.WithContext(context.Background()),
		s.OpenSearchClient.Search.WithIndex("users"),
		s.OpenSearchClient.Search.WithBody(io.NopCloser(io.Reader(bytes.NewBuffer(body)))),
	)
	if err != nil {
		log.Printf("Error searching OpenSearch: %v", err)
		http.Error(w, "Error performing search", http.StatusInternalServerError)
		return
	}
	defer res.Body.Close()

	responseBody, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		http.Error(w, "Error reading search results", http.StatusInternalServerError)
		return
	}

	if res.StatusCode > 299 {
		log.Printf("OpenSearch error: %s", string(responseBody))
		http.Error(w, "Error from search service", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(responseBody)
	if err != nil {
		log.Printf("Error writing response: %v", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
