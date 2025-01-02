package server

import (
	"bytes"
	"io"
	"log"
	"net/http"
)

func printRequestDetails(r *http.Request) {
	log.Println("Headers:")
	for key, values := range r.Header {
		for _, value := range values {
			log.Printf("%s: %s\n", key, value)
		}
	}

	log.Println("Body:")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Println("Error reading body:", err)
		return
	}
	log.Println(string(body))
	r.Body = io.NopCloser(bytes.NewReader(body))
}
