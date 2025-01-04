/***************************************************************************
 * File Name: user-service/server/database.go
 * Author: Bryan SebaRaj
 * Description: Helper functions to connect to PSQL DB
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
	"fmt"
	"log"
	"net"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

func GetEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func ConnectToDB() *sql.DB {
	dbUser := GetEnv("DB_USERNAME", "localtest")
	dbPassword := GetEnv("DB_PASSWORD", "localtest")
	dbEndpoint := GetEnv("DB_ENDPOINT", "localhost")
	dbPort := GetEnv("DB_PORT", "5432")
	dbName := GetEnv("DB_NAME", "my_database")
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
