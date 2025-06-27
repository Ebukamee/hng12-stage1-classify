package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

// --- Structs to define the JSON body for the DHL request ---

type RateRequest struct {
	CustomerDetails    CustomerDetails `json:"customerDetails"`
	Accounts           []Account       `json:"accounts"`
	ProductsAndServices []ProductAndService `json:"productsAndServices,omitempty"` // Optional
	PayerCountryCode   string          `json:"payerCountryCode,omitempty"`      // Optional
	PlannedShippingDateAndTime string   `json:"plannedShippingDateAndTime"`
	UnitOfMeasurement  string          `json:"unitOfMeasurement"`
	IsCustomsDeclarable bool           `json:"isCustomsDeclarable"`
	Packages           []Package       `json:"packages"`
}

type CustomerDetails struct {
	ShipperDetails  ShipperDetails  `json:"shipperDetails"`
	ReceiverDetails ReceiverDetails `json:"receiverDetails"`
}

type ShipperDetails struct {
	PostalCode  string `json:"postalCode"`
	CityName    string `json:"cityName"`
	CountryCode string `json:"countryCode"`
}

type ReceiverDetails struct {
	PostalCode  string `json:"postalCode"`
	CityName    string `json:"cityName"`
	CountryCode string `json:"countryCode"`
}

type Account struct {
	TypeCode string `json:"typeCode"`
	Number   string `json:"number"`
}

type ProductAndService struct {
	ProductCode string `json:"productCode"`
}

type Package struct {
	Weight float64 `json:"weight"`
	Dimensions Dimensions `json:"dimensions"`
}

type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Warning: Environment variable %s not set, using fallback.", key)
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
	fmt.Fprintln(w, `<p>To test the DHL rates endpoint, make a POST request to <strong>/test-dhl-rates</strong>.</p>`)
}

// dhlApiHandler handles the logic for calling the DHL API.
func dhlApiHandler(w http.ResponseWriter, r *http.Request) {

	// We only accept POST requests now
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, please use POST.", http.StatusMethodNotAllowed)
		return
	}
	
	log.Println("Received POST request for /test-dhl-rates...")

	// --- Get credentials securely from environment variables ---
	username := os.Getenv("DHL_USERNAME")
	password := os.Getenv("DHL_PASSWORD")
	// IMPORTANT: You must get your DHL Account Number and add it to your Render Environment Variables
	accountNumber := os.Getenv("DHL_ACCOUNT_NUMBER")

	if username == "" || password == "" || accountNumber == "" {
		errorMsg := "Server configuration error: DHL_USERNAME, DHL_PASSWORD, or DHL_ACCOUNT_NUMBER not set."
		http.Error(w, errorMsg, http.StatusInternalServerError)
		log.Println("FATAL:", errorMsg)
		return
	}
	
	// --- Create the request body with all mandatory parameters ---
	// This is an example for a domestic shipment within Nigeria (Lagos to Abuja)
	// You will need to change these values based on the actual shipment.
	requestBody := RateRequest{
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalCode:  "100001",
				CityName:    "Lagos",
				CountryCode: "NG",
			},
			ReceiverDetails: ReceiverDetails{
				PostalCode:  "900001",
				CityName:    "Abuja",
				CountryCode: "NG",
			},
		},
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   accountNumber,
			},
		},
		// The API requires a future date. We'll use tomorrow's date.
		PlannedShippingDateAndTime: time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05") + " GMT+01:00",
		UnitOfMeasurement:  "metric",
		IsCustomsDeclarable: false, // For a domestic shipment
		Packages: []Package{
			{
				Weight: 2.5, // in kilograms
				Dimensions: Dimensions{
					Length: 30, // in centimeters
					Width:  20,
					Height: 10,
				},
			},
		},
	}

	// Convert the Go struct to a JSON byte slice
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Failed to create JSON body", http.StatusInternalServerError)
		log.Println("Error marshalling JSON:", err)
		return
	}

	// --- Make the POST request to DHL ---
	baseURL := "https://express.api.dhl.com/mydhlapi/test"
	endpoint := "/rates"
	client := &http.Client{}

	// Note: We now use http.NewRequest to specify the POST method and include the body
	req, err := http.NewRequest("POST", baseURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "Failed to create DHL request", http.StatusInternalServerError)
		log.Println("Error creating DHL POST request:", err)
		return
	}

	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Content-Type", "application/json") // Set content type for POST request

	log.Println("Calling DHL API with POST request...")
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

	log.Printf("Successfully served response from DHL. Status: %s", resp.Status)

	// Return the response from DHL back to the user
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	port := getEnv("PORT", "8080")

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/test-dhl-rates", dhlApiHandler)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
