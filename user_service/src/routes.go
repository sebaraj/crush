package main

import (
	"database/sql"
	"net/http"
)

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		DB: db,
	}
}

func (s *Server) initializeRoutes(router *http.ServeMux) {
	router.HandleFunc("/v1/user/info/", s.corsMiddleware(s.handleUser))
	router.HandleFunc("/v1/user/search/", s.corsMiddleware(s.handleSearch))
	router.HandleFunc("/v1/user/picture/", s.corsMiddleware(s.handlePicture))
}
