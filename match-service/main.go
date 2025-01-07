/***************************************************************************
 * File Name: match-service/main.go
 * Author: Bryan SebaRaj
 * Description: Entrypoint for match service pod; initializes server and its dependencies/
 * connections to PostgreSQL (DMS) and SQS.
 * Date Created: 01-07-2025
 *
 * Copyright (c) 2025 Bryan SebaRaj. All rights reserved.
 *
 * License:
 * This file is part of Crush. See the LICENSE file for details.
 ***************************************************************************/

package main

import (
	"context"
	// "crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	// "github.com/aws/aws-sdk-go/aws"
	// "github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	_ "github.com/lib/pq"

	"github.com/sebaraj/crush/match-service/server"
)

func main() {
	// connect to postgresql (RDS)
	db := server.ConnectToDB()

	// initialize SQS client
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("failed to load AWS config: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	queueURL := server.GetEnv("MATCH_QUEUE_URL", "")
	if queueURL == "" {
		log.Fatal("MATCH_QUEUE_URL not set")
	}

	// initialize server
	app := server.NewServer(db, queueURL, sqsClient)
	router := http.NewServeMux()
	app.InitializeRoutes(router)

	server := &http.Server{
		Addr:    ":7000",
		Handler: router,
	}

	// graceful shutdown when testing locally
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("Starting server on :7000")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	<-stop
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}
	if err := app.DB.Close(); err != nil {
		log.Fatalf("Error closing DB: %v", err)
	}
	log.Println("Server shutdown")
}
