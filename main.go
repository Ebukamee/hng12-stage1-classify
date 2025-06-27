package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
)

func main() {
	// Set credentials
	username := "apZ1zO7sH0wB2m"
	password := "S!3yL@3tV!7vH$7x"

	// Build basic auth header
	auth := username + ":" + password
	authHeader := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))

	// Make the request
	req, err := http.NewRequest("GET", "https://express.api.dhl.com/mydhlapi/test", nil)
	if err != nil {
		panic(err)
	}

	// Set the Authorization header
	req.Header.Add("Authorization", authHeader)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Read the response
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Status:", resp.Status)
	fmt.Println("Response:", string(body))
}
