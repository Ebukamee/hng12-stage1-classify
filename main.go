package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		username := "apZ1zO7sH0wB2m"
		password := "S!3yL@3tV!7vH$7x"

		// Encode credentials for Basic Auth
		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))

		// DHL endpoint (NOT /test)
		url := "https://express.api.dhl.com/mydhlapi/rates"

		// Sample request body for DHL /rates (you can adjust this)
		jsonBody := []byte(`{
			"customerDetails": {
				"shipperDetails": {
					"postalCode": "10115",
					"cityName": "Berlin",
					"countryCode": "DE"
				},
				"receiverDetails": {
					"postalCode": "20095",
					"cityName": "Hamburg",
					"countryCode": "DE"
				}
			},
			"plannedShippingDateAndTime": "2025-07-01T12:00:00GMT+01:00",
			"unitOfMeasurement": "metric",
			"isCustomsDeclarable": false,
			"content": "documents",
			"declaredValue": 100,
			"declaredValueCurrency": "EUR",
			"packages": [
				{
					"typeCode": "BOX",
					"weight": 2,
					"dimensions": {
						"length": 10,
						"width": 10,
						"height": 10
					}
				}
			]
		}`)

		// Create the request
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}

		// Set headers
		req.Header.Add("Authorization", "Basic "+auth)
		req.Header.Set("Content-Type", "application/json")

		// Send request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error contacting DHL API", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Read and forward response
		body, _ := io.ReadAll(resp.Body)
		fmt.Println("DHL Response:", string(body)) // logs to Render
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	})

	fmt.Println("✅ Server running on port", port)
	http.ListenAndServe(":"+port, nil)
}
