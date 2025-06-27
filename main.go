package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os" // Import the 'os' package to read environment variables
)

// getEnv gets an environment variable or returns a default value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// rootHandler provides instructions when someone visits the main page.
func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>DHL API Server is running!</h1>")
	fmt.Fprintln(w, `<p>To test the DHL rates endpoint, append <strong>/test-dhl-rates</strong> to the URL.</p>`)
}

// dhlApiHandler handles the logic for calling the DHL API.
func dhlApiHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Received request for /test-dhl-rates...")

	// --- Get credentials securely from environment variables ---
	username := os.Getenv("DHL_USERNAME")
	password := os.Getenv("DHL_PASSWORD")

	if username == "" || password == "" {
		http.Error(w, "Server configuration error: DHL credentials not set.", http.StatusInternalServerError)
		log.Println("FATAL: DHL_USERNAME or DHL_PASSWORD environment variables are not set.")
		return
	}

	baseURL := "https://express.api.dhl.com/mydhlapi/test"
	endpoint := "/rates"

	client := &http.Client{}
	req, err := http.NewRequest("GET", baseURL+endpoint, nil)
	if err != nil {
		http.Error(w, "Failed to create DHL request", http.StatusInternalServerError)
		log.Println("Error creating DHL request:", err)
		return
	}

	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Add("Accept", "application/json")

	log.Println("Calling DHL API...")
	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "Failed to send request to DHL", http.StatusInternalServerError)
		log.Println("Error sending request to DHL:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read DHL response", http.StatusInternalServerError)
		log.Println("Error reading DHL response:", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	log.Printf("Successfully served response from DHL. Status: %s", resp.Status)
	w.Write(body)
}

func main() {
	// --- Use the PORT environment variable provided by Render ---
	port := getEnv("PORT", "8080") // Fallback to 8080 for local testing

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/test-dhl-rates", dhlApiHandler)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
