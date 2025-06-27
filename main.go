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

// --- Structs for the /shipments Request Body ---

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
	fmt.Fprintln(w, "<h1>DHL Shipment API Server</h1>")
	fmt.Fprintln(w, `<p>To create a shipment and get a tracking number, make a POST request to <strong>/create-shipment</strong>.</p>`)
}

// dhlShipmentHandler handles the logic for creating a DHL shipment.
func dhlShipmentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method, please use POST.", http.StatusMethodNotAllowed)
		return
	}

	log.Println("Received POST request for /create-shipment...")

	// --- Get credentials securely from environment variables ---
	username := os.Getenv("DHL_USERNAME")
	password := os.Getenv("DHL_PASSWORD")
	accountNumber := os.Getenv("DHL_ACCOUNT_NUMBER")

	if username == "" || password == "" || accountNumber == "" {
		errorMsg := "Server configuration error: DHL_USERNAME, DHL_PASSWORD, or DHL_ACCOUNT_NUMBER not set."
		http.Error(w, errorMsg, http.StatusInternalServerError)
		log.Println("FATAL:", errorMsg)
		return
	}

	// --- Create the request body for a new shipment ---
	// This is an example for a domestic shipment within Nigeria.
	// Change these details as needed for your specific shipments.
	requestBody := ShipmentRequest{
		PlannedShippingDateAndTime: time.Now().AddDate(0, 0, 1).Format("2006-01-02T15:04:05") + " GMT+01:00",
		Pickup: Pickup{
			IsRequested: false, // Set to true if you want DHL to schedule a pickup
		},
		ProductCode: "N", // 'N' is a common code for Domestic Express
		Accounts: []Account{
			{
				TypeCode: "shipper",
				Number:   accountNumber,
			},
		},
		CustomerDetails: CustomerDetails{
			ShipperDetails: ShipperDetails{
				PostalAddress: PostalAddress{
					PostalCode:  "100001",
					CityName:    "Lagos",
					CountryCode: "NG",
					AddressLine1: "123 Shipper Street",
				},
				ContactInformation: ContactInformation{
					Phone:      "08012345678",
					CompanyName: "Shipper Inc",
					FullName:    "John Shipper",
				},
			},
			ReceiverDetails: ReceiverDetails{
				PostalAddress: PostalAddress{
					PostalCode:  "900001",
					CityName:    "Abuja",
					CountryCode: "NG",
					AddressLine1: "456 Receiver Avenue",
				},
				ContactInformation: ContactInformation{
					Phone:      "09012345678",
					CompanyName: "Receiver Corp",
					FullName:    "Jane Receiver",
				},
			},
		},
		Content: Content{
			Packages: []Package{
				{
					Weight: 1.5,
					Dimensions: Dimensions{
						Length: 20,
						Width:  15,
						Height: 10,
					},
				},
			},
			IsCustomsDeclarable: false,
			Description:         "Business Documents",
			UnitOfMeasurement:   "metric",
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		http.Error(w, "Failed to create JSON body", http.StatusInternalServerError)
		log.Println("Error marshalling JSON:", err)
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

	log.Printf("Calling DHL /shipments endpoint...")
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
	w.Write(body) // Return the full response from DHL
}

func main() {
	port := getEnv("PORT", "8080")

	http.HandleFunc("/", rootHandler)
	// New endpoint for creating shipments
	http.HandleFunc("/create-shipment", dhlShipmentHandler)

	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
