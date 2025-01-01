package main

import (
	"encoding/json"
	// "io"
	// "database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	// "github.com/lib/pq"
)

const numQuestions = 12

type answers struct {
	Answers []int `json:"answers"`
}

func (s *Server) handleAnswers(w http.ResponseWriter, r *http.Request) {
	printRequestDetails(r)
	email := r.URL.Path[len("/v1/user/answers/"):]
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
	case http.MethodPut:
		s.handleUpdateAnswers(w, r, email)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleUpdateAnswers(w http.ResponseWriter, r *http.Request, email string) {
	log.Printf("PUT request to update user answers: %s", email)

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
		if field == "email" {
			continue
		}

		// prevent SQL injection... more rigorous way?
		switch field {
		case "question1", "question2", "question3", "question4", "question5", "question6", "question7", "question8", "question9", "question10", "question11", "question12":
			updateFields = append(updateFields, field+" = $"+fmt.Sprint(i))
			if value == nil {
				http.Error(w, "Invalid value for field: "+field, http.StatusBadRequest)
				log.Printf("Invalid value for field: %s", field)
				return
			}

			var numericValue int
			switch v := value.(type) {
			case string:
				parsedValue, err := strconv.Atoi(v)
				if err != nil {
					http.Error(w, "Invalid value for field: "+field, http.StatusBadRequest)
					log.Printf("Invalid value for field (not a number): %s", field)
					return
				}
				numericValue = parsedValue
			case int:
				numericValue = v
			case float64:
				numericValue = int(v)
			default:
				http.Error(w, "Invalid value for field: "+field, http.StatusBadRequest)
				log.Printf("Invalid value for field (unsupported type): %s", field)
				return
			}

			if numericValue < 0 || numericValue > 5 {
				http.Error(w, "Invalid value for field: "+field, http.StatusBadRequest)
				log.Printf("Invalid value for field (out of range): %s", field)
				return
			}

			updateValues = append(updateValues, numericValue)
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
	query := "UPDATE answers SET " + joinFields(updateFields, ", ") + " WHERE email = $" + fmt.Sprint(i)

	_, err = tx.Exec(query, updateValues...)
	if err != nil {
		http.Error(w, "Failed to update user answers", http.StatusInternalServerError)
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
	_, err = w.Write([]byte("User answers updated successfully"))
	if err != nil {
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
	log.Printf("User answers updated successfully: %s", email)
}
