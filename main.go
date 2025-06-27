package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	// Use the platform-assigned port (e.g., for Render, Railway, Heroku)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // fallback for local dev
	}

	// Define a basic route
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		username := "apZ1zO7sH0wB2m"
		password := "S!3yL@3tV!7vH$7x"

		auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
		req, err := http.NewRequest("GET", "https://express.api.dhl.com/mydhlapi/test", nil)
		if err != nil {
			http.Error(w, "Error creating request", http.StatusInternalServerError)
			return
		}

		req.Header.Add("Authorization", "Basic "+auth)

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			http.Error(w, "Error calling DHL API", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)

		// Pass the API response to the browser
		io.Copy(w, resp.Body)
	})

	fmt.Println("🚀 Server is running on port", port)
	err := http.ListenAndServe(":"+port, nil)
	if err != nil {
		panic(err)
	}
}
