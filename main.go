package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// dhlApiHandler is the function that will handle incoming web requests.
func dhlApiHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received request for /test-dhl-rates...")

	// DHL API Credentials and Test URL
	username := "apZ1zO7sH0wB2m"
	password := "S!3yL@3tV!7vH$7x"
	baseURL := "https://express.api.dhl.com/mydhlapi/test"
	endpoint := "/rates"

	// Create a new HTTP client
	client := &http.Client{}

	// Create a new GET request to the DHL API
	req, err := http.NewRequest("GET", baseURL+endpoint, nil)
	if err != nil {
		http.Error(w, "Failed to create DHL request", http.StatusInternalServerError)
		log.Println("Error creating DHL request:", err)
		return
	}

	// Create and set the Basic Authentication header
	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Add("Accept", "application/json") // It's good practice to specify you accept JSON

	// Send the request to DHL
	fmt.Println("Calling DHL API...")
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request to DHL", http.StatusInternalServerError)
		log.Println("Error sending request to DHL:", err)
		return
	}
	defer resp.Body.Close()

	// Read the response from DHL
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read DHL response", http.StatusInternalServerError)
		log.Println("Error reading DHL response:", err)
		return
	}

	// Set the content type of our response to the user
	w.Header().Set("Content-Type", "application/json")

	// Write the DHL response back to the user who visited our server
	fmt.Fprintf(w, "Response from DHL API:\n\nStatus: %s\nBody: %s", resp.Status, string(body))
	log.Println("Successfully served response from DHL. Status:", resp.Status)
}

func main() {
	// Register our handler function to respond to requests at the "/test-dhl-rates" URL path
	http.HandleFunc("/test-dhl-rates", dhlApiHandler)

	// Define the port the server will listen on
	port := "8080"

	// Start the server
	fmt.Printf("Server starting on port %s...\n", port)
	fmt.Printf("Visit http://localhost:%s/test-dhl-rates to trigger the DHL API call.\n", port)

	// The ListenAndServe function blocks forever, or until an error occurs.
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
