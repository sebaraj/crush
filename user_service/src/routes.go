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
	router.HandleFunc("/v1/user/info/", s.handleUser)
	router.HandleFunc("/v1/user/search/", s.handleSearch)
	router.HandleFunc("/v1/user/picture/", s.handlePicture)
}
