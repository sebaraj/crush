package main

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func connectToDB() *sql.DB {
	dbUser := getEnv("DB_USERNAME", "localtest")
	dbPassword := getEnv("DB_PASSWORD", "localtest")
	dbEndpoint := getEnv("DB_ENDPOINT", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbName := getEnv("DB_NAME", "my_database")
	fmt.Println("DB_USER:", dbUser)
	fmt.Println("DB_PASSWORD:", dbPassword)

	if dbUser == "" || dbPassword == "" || dbEndpoint == "" || dbPort == "" || dbName == "" {
		log.Fatal("One or more required environment variables are missing")
	}
	dbEndpoint = strings.Split(dbEndpoint, ":")[0]

	ips, err := net.LookupIP(dbEndpoint)
	if err != nil {
		log.Fatalf("Failed to resolve hostname: %v", err)
	}

	if len(ips) == 0 {
		log.Fatalf("No IP addresses found for hostname: %s", dbEndpoint)
	}

	dbIP := ips[0].String()
	log.Printf("Resolved %s to %s", dbEndpoint, dbIP)

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		dbIP, dbPort, dbUser, dbPassword, dbName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("Unable to connect to DB: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("Unable to reach DB: %v", err)
	}
	log.Println("Successfully connected to the database")
	return db
}
