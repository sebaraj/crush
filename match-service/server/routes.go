/***************************************************************************
 * File Name: match-service/server/routes.go
 * Author: Bryan SebaRaj
 * Description: Define server struct and routes
 * Date Created: 01-07-2025
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

	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type Server struct {
	DB         *sql.DB
	SQS_URL    string
	SQS_Client *sqs.Client
}

func NewServer(db *sql.DB, sqs_url string, sqs_client *sqs.Client) *Server {
	return &Server{
		DB:         db,
		SQS_URL:    sqs_url,
		SQS_Client: sqs_client,
	}
}

func (s *Server) InitializeRoutes(router *http.ServeMux) {
	router.HandleFunc("GET /v1/match/", s.corsMiddleware(s.HandleGetMatch)) // gets all matches belonging to user
	// router.HandleFunc("PUT /v1/match/", s.corsMiddleware(s.HandleUpdateMatch)) // updates match status
}
