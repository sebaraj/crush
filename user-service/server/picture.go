package server

import (
	"log"
	"net/http"
	"time"
	// "encoding/json"
	// "io"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
)

func (s *Server) HandlePicture(w http.ResponseWriter, r *http.Request) {
	printRequestDetails(r)
	userEmail := r.URL.Path[len("/v1/user/picture/"):]
	log.Printf("Search: %s", userEmail)

	emailFromToken, err := s.validateOAuthToken(r)
	if err != nil || userEmail != emailFromToken {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		s.handleS3URLRequest(w, r, userEmail)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleS3URLRequest(w http.ResponseWriter, r *http.Request, userEmail string) {
	log.Printf("GET request for S3 signed url: %s", userEmail)

	// generate S3 signed URL
	objectKey := "user-images/" + userEmail + ".jpg"
	expires := 5 * time.Minute

	req, output := s.S3Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(s.S3Bucket),
		Key:    aws.String(objectKey),
	})
	log.Printf("req: %v", req)
	log.Printf("output: %v", output)
	url, err := req.Presign(expires)
	if err != nil {
		log.Printf("Failed to sign request: %v", err)
		http.Error(w, "Failed to sign request", http.StatusInternalServerError)
		return
	}

	// update user's picture S3 URL in database
	_, err = s.DB.Exec("UPDATE users SET picture_s3_url = $1 WHERE email = $2", url, userEmail)
	if err != nil {
		log.Printf("Failed to update user's picture S3 URL: %v", err)
		http.Error(w, "Failed to update user's picture S3 URL", http.StatusInternalServerError)
		return
	}

	// return S3 signed URL to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(`{"url": "` + url + `"}`))
	if err != nil {
		log.Printf("Failed to write response: %v", err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
