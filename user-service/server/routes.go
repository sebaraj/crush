/***************************************************************************
 * File Name: user-service/server/routes.go
 * Author: Bryan SebaRaj
 * Description: Define server struct and routes
 * Date Created: 01-01-2025
 *
 * Copyright (c) 2025 Bryan SebaRaj. All rights reserved.
 *
 * License:
 * This file is part of Crush. See the LICENSE file for details.
 ***************************************************************************/

package server

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

func (s *Server) InitializeRoutes(router *http.ServeMux) {
	router.HandleFunc("/v1/user/info/", s.corsMiddleware(s.HandleUser))
	router.HandleFunc("/v1/user/answers/", s.corsMiddleware(s.HandleAnswers))
	router.HandleFunc("/v1/user/search/", s.corsMiddleware(s.HandleSearch))
	router.HandleFunc("/v1/user/picture/", s.corsMiddleware(s.HandlePicture))
}
