/***************************************************************************
 * File Name: user-service/server/user.go
 * Author: Bryan SebaRaj
 * Description: Handler for getting and updating user info; defines User struct
 * Date Created: 01-01-2025
 *
 * Copyright (c) 2025 Bryan SebaRaj. All rights reserved.
 *
 * License:
 * This file is part of Crush. See the LICENSE file for details.
 ***************************************************************************/

package server

import (
	"encoding/json"
	// "io"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	// "github.com/lib/pq"
)

type User struct {
	Email              string   `json:"email"`
	IsActive           bool     `json:"is_active"`
	Name               string   `json:"name"`
	ResidentialCollege string   `json:"residential_college"`
	NotifPref          bool     `json:"notif_pref"`
	GraduatingYear     int      `json:"graduating_year"`
	Gender             int      `json:"gender"`
	PartnerGenders     int      `json:"partner_genders"`
	Instagram          string   `json:"instagram"`
	Snapchat           string   `json:"snapchat"`
	PhoneNumber        string   `json:"phone_number"`
	PictureS3URL       string   `json:"picture_s3_url"`
	Interests          []string `json:"interests"`
	Answers            []int    `json:"answers"`
}

const (
	NumInterests = 5
	NumQuestions = 12
)

func (s *Server) HandleUser(w http.ResponseWriter, r *http.Request) {
	printRequestDetails(r)
	email := r.URL.Path[len("/v1/user/info/"):]
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

	switch r.Method {
	case http.MethodGet:
		s.handleGetUser(w, r, email)
	case http.MethodPut:
		s.handleUpdateUser(w, r, email)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("GET request for user: %s", email)

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

	var result User
	var name sql.NullString
	var residentialCollege sql.NullString
	var graduatingYear sql.NullInt64
	var gender sql.NullInt64
	var partnerGenders sql.NullInt64
	var instagram sql.NullString
	var snapchat sql.NullString
	var phoneNumber sql.NullString
	var pictureS3URL sql.NullString
	var interests [NumInterests]sql.NullString
	var answers [NumQuestions]sql.NullInt64

	query := `
		SELECT 
			u.email, 
			u.is_active, 
			u.name, 
			u.residential_college, 
      u.notif_pref,
			u.graduating_year, 
			u.gender, 
			u.partner_genders, 
			u.instagram, 
			u.snapchat, 
			u.phone_number, 
			u.picture_s3_url, 
			u.interest_1, 
			u.interest_2, 
			u.interest_3, 
			u.interest_4, 
			u.interest_5,
			a.question1, 
			a.question2, 
			a.question3, 
			a.question4, 
			a.question5, 
			a.question6, 
			a.question7, 
			a.question8, 
			a.question9, 
			a.question10, 
			a.question11, 
			a.question12
		FROM users u
		LEFT JOIN answers a ON u.email = a.email
		WHERE u.email = $1
	`
	row := tx.QueryRow(query, email)

	err = row.Scan(
		&result.Email,
		&result.IsActive,
		&name,
		&residentialCollege,
		&result.NotifPref,
		&graduatingYear,
		&gender,
		&partnerGenders,
		&instagram,
		&snapchat,
		&phoneNumber,
		&pictureS3URL,
		&interests[0],
		&interests[1],
		&interests[2],
		&interests[3],
		&interests[4],
		&answers[0],
		&answers[1],
		&answers[2],
		&answers[3],
		&answers[4],
		&answers[5],
		&answers[6],
		&answers[7],
		&answers[8],
		&answers[9],
		&answers[10],
		&answers[11],
	)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to query database", http.StatusInternalServerError)
		log.Printf("Failed to query database: %v", err)
		return
	}

	result.ResidentialCollege = getStringValue(residentialCollege)
	result.Name = getStringValue(name)
	result.GraduatingYear = getIntValue(graduatingYear)
	result.Gender = getIntValue(gender)
	result.PartnerGenders = getIntValue(partnerGenders)
	result.Instagram = getStringValue(instagram)
	result.Snapchat = getStringValue(snapchat)
	result.PhoneNumber = getStringValue(phoneNumber)
	result.PictureS3URL = getStringValue(pictureS3URL)
	result.Interests = filterNullStrings(interests[:])
	result.Answers = filterNullInts(answers[:])

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit database transaction", http.StatusInternalServerError)
		log.Printf("Failed to commit database transaction: %v", err)
		return
	}

	jsonResponse, err := json.Marshal(result)
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

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("PUT request to update user: %s", email)

	var updates map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&updates)
	if err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		log.Printf("Failed to decode request body: %v", err)
		return
	}

	tx, err := s.DB.Begin()
	if err != nil {
		http.Error(w, "Failed to start database transaction", http.StatusInternalServerError)
		log.Printf("Failed to start database transaction: %v", err)
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			log.Printf("Transaction rolled back: %v", err)
		}
	}()

	updateFields := []string{}
	updateValues := []interface{}{}
	i := 1

	for field, value := range updates {
		if field == "email" || field == "is_active" || field == "picture_s3_url" {
			continue
		}

		// prevent SQL injection... more rigorous way?
		switch field {
		case "name", "residential_college", "graduating_year", "gender",
			"partner_genders", "instagram", "snapchat", "phone_number",
			"interest_1", "interest_2", "interest_3", "interest_4", "interest_5", "notif_pref":
			updateFields = append(updateFields, field+" = $"+fmt.Sprint(i))
			updateValues = append(updateValues, value)
			i++
		default:
			http.Error(w, "Invalid field in request body: "+field, http.StatusBadRequest)
			log.Printf("Invalid field in request body: %s", field)
			return
		}
	}

	if len(updateFields) == 0 {
		http.Error(w, "No valid fields to update", http.StatusBadRequest)
		return
	}

	updateValues = append(updateValues, email)
	query := "UPDATE users SET " + joinFields(updateFields, ", ") + " WHERE email = $" + fmt.Sprint(i)

	_, err = tx.Exec(query, updateValues...)
	if err != nil {
		http.Error(w, "Failed to update user", http.StatusInternalServerError)
		log.Printf("Failed to execute update query: %v", err)
		return
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit transaction", http.StatusInternalServerError)
		log.Printf("Failed to commit transaction: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	log.Printf("User updated successfully: %s", email)
}
