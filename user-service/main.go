package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	_ "github.com/lib/pq"

	"github.com/opensearch-project/opensearch-go"
	"github.com/sebaraj/crush/user-service/server"
)

func main() {
	db := server.ConnectToDB()

	// connect to s3
	s3Region := server.GetEnv("S3_REGION", "")
	s3Bucket := server.GetEnv("S3_BUCKET", "")

	if s3Region == "" || s3Bucket == "" {
		log.Fatal("One or more required environment variables for S3 are missing")
		return
	}

	// connect to opensearch
	opensearchEndpoint := server.GetEnv("OPENSEARCH_ENDPOINT", "")
	osClient, err := opensearch.NewClient(opensearch.Config{
		Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}},
		Addresses: []string{opensearchEndpoint},
	})
	log.Printf("OpenSearch endpoint%s", opensearchEndpoint)

	if err != nil {
		log.Fatalf("Error creating the OpenSearch client: %s", err)
	}
	log.Printf("OpenSearch client created")

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(s3Region),
	}))

	app := server.NewServer(db, s3Bucket, s3Region, s3.New(sess), osClient)
	router := http.NewServeMux()
	app.InitializeRoutes(router)

	server := &http.Server{
		Addr:    ":6000",
		Handler: router,
	}

	// graceful shutdown when testing locally
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		log.Println("Starting server on :6000")
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
