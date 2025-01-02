package server

import (
	"context"
	"errors"
	"net/http"
	"os"

	"google.golang.org/api/idtoken"
)

// middleware to validate OAuth token using client key. use on any protected routes
// returns (email|"", nil|error)
func (s *Server) validateOAuthToken(r *http.Request) (string, error) {
	token := r.Header.Get("Authorization")
	if token == "" {
		return "", errors.New("no token provided")
	}

	oauthClient := os.Getenv("OAUTH_CLIENT")
	if oauthClient == "" {
		return "", errors.New("OAUTH_CLIENT not set")
	}
	ctx := context.Background()
	payload, err := idtoken.Validate(ctx, token, oauthClient)
	if err != nil {
		return "", err
	}

	email, ok := payload.Claims["email"].(string)
	if !ok {
		return "", errors.New("email not found in token")
	}

	return email, nil
}

func (s *Server) corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // change to yalecrush.com for prod
		w.Header().Set("Access-Control-Allow-Methods", "GET, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}
