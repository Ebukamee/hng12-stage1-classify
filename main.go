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

		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		url := "https://express.api.dhl.com/mydhlapi/shipments"

		// Sample DHL test shipment payload
		jsonBody := []byte(`{
			"plannedShippingDateAndTime": "2025-07-01T12:00:00GMT+01:00",
			"pickup": {
				"isRequested": false
			},
			"productCode": "P",
			"customerDetails": {
				"shipperDetails": {
					"postalCode": "10115",
					"cityName": "Berlin",
					"countryCode": "DE",
					"addressLine1": "Shipper Street 1",
					"name": "Test Shipper",
					"email": "shipper@example.com",
					"phone": "1234567890"
				},
				"receiverDetails": {
					"postalCode": "20095",
					"cityName": "Hamburg",
					"countryCode": "DE",
					"addressLine1": "Receiver Street 1",
					"name": "Test Receiver",
					"email": "receiver@example.com",
					"phone": "0987654321"
				}
			},
			"content": "documents",
			"packages": [
				{
					"typeCode": "BOX",
					"weight": 1.5,
					"dimensions": {
						"length": 10,
						"width": 10,
						"height": 10
					}
				}
			]
		}`)

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			http.Error(w, "Failed to create request", http.StatusInternalServerError)
			return
		}

		req.Header.Add("Authorization", "Basic "+auth)
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "DHL request failed", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		body, _ := io.ReadAll(resp.Body)
		fmt.Println("DHL Shipment Response:", string(body))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		w.Write(body)
	})

	fmt.Println("🚀 Server running on port", port)
	http.ListenAndServe(":"+port, nil)
}
