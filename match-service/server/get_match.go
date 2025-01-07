/***************************************************************************
 * File Name: match-service/server/user.go
 * Author: Bryan SebaRaj
 * Description: Handler for getting match info; defines match struct
 * Date Created: 01-07-2025
 *
 * Copyright (c) 2025 Bryan SebaRaj. All rights reserved.
 *
 * License:
 * This file is part of Crush. See the LICENSE file for details.
 ***************************************************************************/

package server

import (
	// "database/sql"
	"encoding/json"
	// "fmt"
	"log"
	"net/http"
)

type Match struct {
	SourceEmail      string `json:"source_email"`
	TargetEmail      string `json:"target_email"`
	SourceInterested bool   `json:"source_interested"`
	TargetInterested bool   `json:"target_interested"`
	ServerGenerated  bool   `json:"server_generated"`
	Week             string `json:"week"`
}

func (s *Server) HandleGetMatch(w http.ResponseWriter, r *http.Request) {
	printRequestDetails(r)
	email := r.URL.Path[len("/v1/match/"):]
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	log.Printf("Email: %s", email)
	emailFromToken, err := s.validateOAuthToken(r)
	if err != nil || emailFromToken != email {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	log.Printf("GET request for matches for user: %s", email)

	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start database transaction", http.StatusInternalServerError)
		log.Printf("Failed to start database transaction: %v", err)
		return
	}
	// no-op in psql if tx.Commit is called first
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back: %v", err)
		}
	}()

	query := `
		SELECT user1_email, user2_email, user1_interested, user2_interested, server_generated, week FROM matches WHERE user1_email = $1 OR user2_email = $1
	`
	rows, queryErr := tx.Query(query, email)
	if queryErr != nil {
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		log.Printf("Failed to query database: %v", queryErr)
		err = queryErr
		return
	}
	// array of matches
	defer rows.Close()

	var results []Match

	for rows.Next() {
		var result Match
		scanErr := rows.Scan(&result.SourceEmail, &result.TargetEmail, &result.SourceInterested, &result.TargetInterested, &result.ServerGenerated, &result.Week)
		if scanErr != nil {
			http.Error(w, "Failed to scan database results", http.StatusInternalServerError)
			log.Printf("Failed to scan database results: %v", scanErr)
			err = scanErr
			return
		}
		results = append(results, result)
	}
	errNext := rows.Err()
	if errNext != nil {
		http.Error(w, "Error iterating rows", http.StatusInternalServerError)
		log.Printf("Error iterating rows: %v", errNext)
		err = errNext
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit database transaction", http.StatusInternalServerError)
		log.Printf("Failed to commit database transaction: %v", err)
		return
	}

	jsonResponse, err := json.Marshal(results)
	if err != nil {
		http.Error(w, "Failed to marshal JSON response", http.StatusInternalServerError)
		log.Printf("Failed to marshal JSON response: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		log.Printf("Failed to write response: %v", err)
		return
	}
}
