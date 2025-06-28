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
)

// ---------------------- Structs for DHL Shipment ----------------------

type ShipmentRequest struct {
	PlannedShippingDateAndTime string          `json:"plannedShippingDateAndTime"`
	Pickup                     Pickup          `json:"pickup"`
	ProductCode                string          `json:"productCode"`
	Accounts                   []Account       `json:"accounts"`
	CustomerDetails            CustomerDetails `json:"customerDetails"`
	Content                    Content         `json:"content"`
}

type Pickup struct {
	IsRequested bool `json:"isRequested"`
}

type Account struct {
	TypeCode string `json:"typeCode"`
	Number   string `json:"number"`
}

type CustomerDetails struct {
	ShipperDetails  ShipperDetails  `json:"shipperDetails"`
	ReceiverDetails ReceiverDetails `json:"receiverDetails"`
}

type ShipperDetails struct {
	PostalAddress     PostalAddress     `json:"postalAddress"`
	ContactInformation ContactInformation `json:"contactInformation"`
}

type ReceiverDetails struct {
	PostalAddress     PostalAddress     `json:"postalAddress"`
	ContactInformation ContactInformation `json:"contactInformation"`
}

type PostalAddress struct {
	PostalCode   string `json:"postalCode"`
	CityName     string `json:"cityName"`
	CountryCode  string `json:"countryCode"`
	AddressLine1 string `json:"addressLine1"`
	CountyName   string `json:"countyName,omitempty"`
}

type ContactInformation struct {
	Email       string `json:"email,omitempty"`
	Phone       string `json:"phone"`
	MobilePhone string `json:"mobilePhone,omitempty"`
	CompanyName string `json:"companyName"`
	FullName    string `json:"fullName"`
}

type Content struct {
	Packages            []Package `json:"packages"`
	IsCustomsDeclarable bool      `json:"isCustomsDeclarable"`
	Description         string    `json:"description"`
	UnitOfMeasurement   string    `json:"unitOfMeasurement"`
}

type Package struct {
	Weight     float64    `json:"weight"`
	Dimensions Dimensions `json:"dimensions"`
}

type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

// ---------------------- Utility ----------------------

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Printf("Warning: env %s not set, using fallback.", key)
	return fallback
}

// ---------------------- Root Handler ----------------------

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>DHL Live Shipment API Server</h1>")
	fmt.Fprintln(w, `<p>Send a POST request to <strong>/create-shipment</strong> with shipment data in JSON format.</p>`)
}

// ---------------------- DHL Shipment Handler ----------------------

func dhlShipmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, please use POST.", http.StatusMethodNotAllowed)
		return
	}

	log.Println("📦 Received request to create a DHL shipment")

	// Parse incoming shipment JSON
	var requestBody ShipmentRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid JSON body: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Load DHL credentials
	username := os.Getenv("DHL_USERNAME")
	password := os.Getenv("DHL_PASSWORD")
	accountNumber := os.Getenv("DHL_ACCOUNT_NUMBER")

	if username == "" || password == "" || accountNumber == "" {
		http.Error(w, "Missing DHL env variables", http.StatusInternalServerError)
		log.Println("❌ DHL_USERNAME, DHL_PASSWORD, or DHL_ACCOUNT_NUMBER is not set")
		return
	}

	// Inject secure account number into the request
	if len(requestBody.Accounts) == 0 {
		requestBody.Accounts = []Account{
			{TypeCode: "shipper", Number: accountNumber},
		}
	} else {
		for i := range requestBody.Accounts {
			requestBody.Accounts[i].Number = accountNumber
		}
	}

	// Marshal to JSON
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Failed to marshal request", http.StatusInternalServerError)
		log.Println("❌ Marshal error:", err)
		return
	}

	// ✅ LIVE DHL production URL
	baseURL := "https://express.api.dhl.com/mydhlapi"
	endpoint := "/shipments"
	fullURL := baseURL + endpoint

	req, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "Failed to create request to DHL", http.StatusInternalServerError)
		log.Println("❌ Request creation error:", err)
		return
	}

	// Auth
	auth := base64.StdEncoding.EncodeToString([]byte(username + ":" + password))
	req.Header.Add("Authorization", "Basic "+auth)
	req.Header.Set("Content-Type", "application/json")

	log.Println("🚀 Sending live shipment request to DHL...")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		http.Error(w, "Error contacting DHL", http.StatusInternalServerError)
		log.Println("❌ HTTP error:", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read DHL response", http.StatusInternalServerError)
		log.Println("❌ Read error:", err)
		return
	}

	log.Printf("✅ DHL response received with status: %s", resp.Status)

	// Send DHL response back to client
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

// ---------------------- Main Entry Point ----------------------

func main() {
	port := getEnv("PORT", "8080")

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/create-shipment", dhlShipmentHandler)

	log.Printf("🚀 Server running on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("❌ Server startup failed:", err)
	}
}
