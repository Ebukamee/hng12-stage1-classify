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

// --- The Structs remain the same, as they define the contract with the DHL API ---
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
	PostalAddress PostalAddress `json:"postalAddress"`
	ContactInformation ContactInformation `json:"contactInformation"`
}

type ReceiverDetails struct {
	PostalAddress PostalAddress `json:"postalAddress"`
	ContactInformation ContactInformation `json:"contactInformation"`
}

type PostalAddress struct {
	PostalCode  string `json:"postalCode"`
	CityName    string `json:"cityName"`
	CountryCode string `json:"countryCode"`
	AddressLine1 string `json:"addressLine1"`
	CountyName string `json:"countyName,omitempty"`
}

type ContactInformation struct {
	Email      string `json:"email,omitempty"`
	Phone      string `json:"phone"`
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
	Weight float64 `json:"weight"`
	Dimensions Dimensions `json:"dimensions"`
}

type Dimensions struct {
	Length float64 `json:"length"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}


func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintln(w, "<h1>DHL Dynamic Shipment API Server</h1>")
	fmt.Fprintln(w, `<p>To create a shipment, make a POST request to <strong>/create-shipment</strong> with your shipment data in the JSON body.</p>`)
}

// dhlShipmentHandler now acts as a dynamic proxy to the DHL API.
func dhlShipmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, please use POST.", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Received dynamic request for /create-shipment...")

	// --- Decode the JSON body from the incoming request ---
	var requestBody ShipmentRequest
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Bad request: could not decode JSON body. "+err.Error(), http.StatusBadRequest)
		return
	}
	
	// --- Get credentials securely from environment variables ---
	username := os.Getenv("DHL_USERNAME")
	password := os.Getenv("DHL_PASSWORD")

	if username == "" || password == "" {
		errorMsg := "Server configuration error: DHL credentials not set."
		http.Error(w, errorMsg, http.StatusInternalServerError)
		log.Println("FATAL:", errorMsg)
		return
	}

	// Marshal the received body back into JSON format to send to DHL.
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Failed to re-marshal JSON body", http.StatusInternalServerError)
		log.Println("Error re-marshalling JSON:", err)
		return
	}

	// --- Make the POST request to DHL ---
	baseURL := "https://express.api.dhl.com/mydhlapi/test"
	endpoint := "/shipments"
	client := &http.Client{}

	req, err := http.NewRequest("POST", baseURL+endpoint, bytes.NewBuffer(jsonBody))
	if err != nil {
		http.Error(w, "Failed to create DHL request", http.StatusInternalServerError)
		log.Println("Error creating DHL POST request:", err)
		return
	}

	auth := username + ":" + password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))
	req.Header.Add("Authorization", "Basic "+encodedAuth)
	req.Header.Set("Content-Type", "application/json")

	log.Printf("Calling DHL /shipments endpoint with dynamic data...")
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

	log.Printf("Received response from DHL. Status: %s", resp.Status)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(body)
}

func main() {
	port := getEnv("PORT", "8080")

	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/create-shipment", dhlShipmentHandler)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
