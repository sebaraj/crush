/***************************************************************************
 * File Name: match-service/server/put-match.go
 * Author: Bryan SebaRaj
 * Description: Handler for inserting/updating match into SQS
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
	"context"
	"encoding/json"
	// "fmt"
	"log"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type SQSMessage struct {
	EmailSource string `json:"email_source"`
	EmailTarget string `json:"email_target"`
	Date        string `json:"date"`
	WantsMatch  bool   `json:"wants_match"`
}

func (s *Server) HandleUpdateMatch(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
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

	log.Printf("Update request for matches for user: %s", email)

	// parse json request body
	var incomingMatch Match
	if err := json.NewDecoder(r.Body).Decode(&incomingMatch); err != nil {
		http.Error(w, "Failed to decode JSON request body", http.StatusBadRequest)
		log.Printf("Failed to decode JSON request body: %v", err)
		return
	}
	if incomingMatch.SourceEmail != email {
		http.Error(w, "Source email does not match path user", http.StatusBadRequest)
		return
	}

	msg := SQSMessage{
		EmailSource: incomingMatch.SourceEmail,
		EmailTarget: incomingMatch.TargetEmail,
		Date:        incomingMatch.Week,
		WantsMatch:  incomingMatch.SourceInterested || incomingMatch.TargetInterested,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		http.Error(w, "Failed to marshal message", http.StatusInternalServerError)
		log.Printf("Failed to marshal message: %v", err)
		return
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(s.SQS_URL),
		MessageBody: aws.String(string(body)),
	}

	_, err = s.SQS_Client.SendMessage(ctx, input)
	if err != nil {
		http.Error(w, "Failed to send message to SQS", http.StatusInternalServerError)
		log.Printf("Failed to send message to SQS: %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
