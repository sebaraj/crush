/***************************************************************************
 * File Name: auth-service/main.go
 * Author: Bryan SebaRaj
 * Description: Entrypoint for auth service pod; initializes server and its dependencies/
 * connection to PostgreSQL (DMS).
 * Date Created: 01-01-2025
 *
 * Copyright (c) 2025 Bryan SebaRaj. All rights reserved.
 *
 * License:
 * This file is part of Crush. See the LICENSE file for details.
 ***************************************************************************/

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/lib/pq"
	"github.com/sebaraj/crush/auth-service/server"
)

func main() {
	// connect to postgresql (RDS)
	db := server.ConnectToDB()

	// initialize server
	app := server.NewServer(db)
	router := http.NewServeMux()
	router.HandleFunc("/v1/auth", app.CorsMiddleware(app.HandleAuth))

	server := &http.Server{
		Addr:    ":5678",
		Handler: router,
	}

	// graceful shutdown when testing locally
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("Starting server on :5678")
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
