/***************************************************************************
 * File Name: auth-service/server/logging.go
 * Author: Bryan SebaRaj
 * Description: Helper function for logging HTTP request details.
 * Date Created: 01-01-2025
 *
 * Copyright (c) 2025 Bryan SebaRaj. All rights reserved.
 *
 * License:
 * This file is part of Crush. See the LICENSE file for details.
 ***************************************************************************/

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
