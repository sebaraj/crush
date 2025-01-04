package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "github.com/lib/pq"
)

func main() {
	db := connectToDB()
	app := NewServer(db)
	router := http.NewServeMux()
	router.HandleFunc("/v1/auth", app.corsMiddleware(app.handleAuth))

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
