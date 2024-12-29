package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"

	_ "github.com/lib/pq"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		DB: db,
	}
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	var now string
	err := s.DB.QueryRow("SELECT NOW()").Scan(&now)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("DB Query Error:", err)
		return
	}

	response := fmt.Sprintf("pong! Current DB time is: %s", now)
	_, err = w.Write([]byte(response))
	if err != nil {
		log.Println("Error writing response:", err)
	}
}
