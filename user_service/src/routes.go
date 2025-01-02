package main

import (
	"database/sql"
	"net/http"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/opensearch-project/opensearch-go"
)

type Server struct {
	DB               *sql.DB
	S3Bucket         string
	S3Region         string
	S3Client         *s3.S3
	OpenSearchClient *opensearch.Client
}

func NewServer(db *sql.DB, bucket string, s3Region string, s3Client *s3.S3, opensearchClient *opensearch.Client) *Server {
	return &Server{
		DB:               db,
		S3Bucket:         bucket,
		S3Region:         s3Region,
		S3Client:         s3Client,
		OpenSearchClient: opensearchClient,
	}
}

func (s *Server) initializeRoutes(router *http.ServeMux) {
	router.HandleFunc("/v1/user/info/", s.corsMiddleware(s.handleUser))
	router.HandleFunc("/v1/user/answers/", s.corsMiddleware(s.handleAnswers))
	router.HandleFunc("/v1/user/search/", s.corsMiddleware(s.handleSearch))
	router.HandleFunc("/v1/user/picture/", s.corsMiddleware(s.handlePicture))
}
