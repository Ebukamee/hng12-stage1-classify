package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"strconv"
	"unicode"

	"github.com/gorilla/mux"
)

// Number struct for response
type Number struct {
	Integer     int      `json:"number,omitempty"`
	IsPrime     bool     `json:"is_prime,omitempty"`
	IsPerfect   bool     `json:"is_perfect,omitempty"`
	Properties  []string `json:"properties,omitempty"`
	SumOfDigits int      `json:"digit_sum,omitempty"`
	FunFact     string   `json:"fun_fact,omitempty"`
	Error       bool     `json:"error"`
}

// isPrime checks if a number is prime
func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i <= int(math.Sqrt(float64(n))); i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

// isPerfect checks if a number is a perfect number
func isPerfect(n int) bool {
	if n < 2 {
		return false
	}
	sum := 1
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			sum += i
			if i != n/i {
				sum += n / i
			}
		}
	}
	return sum == n
}

// isArmstrong checks if a number is an Armstrong number
func isArmstrong(n int) bool {
	temp, sum, digits := n, 0, 0

	// Count number of digits
	for temp != 0 {
		digits++
		temp /= 10
	}

	temp = n // Reset temp

	// Compute sum of each digit raised to the power of digits
	for temp != 0 {
		digit := temp % 10
		sum += int(math.Pow(float64(digit), float64(digits)))
		temp /= 10
	}

	return sum == n
}

// sumOfDigits calculates the sum of digits of a number
func sumOfDigits(n int) int {
	sum := 0
	for n != 0 {
		sum += n % 10
		n /= 10
	}
	return sum
}

// Properties returns the properties of a number
func Properties(n int) []string {
	properties := []string{}
	if isArmstrong(n) {
		properties = append(properties, "armstrong")
	}
	if n%2 == 0 {
		properties = append(properties, "even")
	} else {
		properties = append(properties, "odd")
	}
	return properties
}

// fetchFunFact fetches a fun fact from Numbers API
func fetchFunFact(n int) string {
	url := fmt.Sprintf("http://numbersapi.com/%d/math", n)
	resp, err := http.Get(url)
	if err != nil {
		return "Could not fetch fun fact"
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "Error reading fun fact"
	}

	return string(body)
}

// API Handler
func get(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query := r.URL.Query().Get("number")

	// Check for empty query
	if query == "" {
		errorResponse := Number{
			Error:   true,
			FunFact: "null",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Check if input is alphabetic
	isAlphabet := true
	for _, ch := range query {
		if !unicode.IsLetter(ch) {
			isAlphabet = false
			break
		}
	}
	if isAlphabet {
		errorResponse := Number{
			Error:   true,
			FunFact: "alphabet",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Try converting to an integer
	num, err := strconv.Atoi(query)
	if err != nil {
		errorResponse := Number{
			Error:   true,
			FunFact: "invalid",
		}
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(errorResponse)
		return
	}

	// Generate response data
	response := Number{
		Integer:     num,
		IsPrime:     isPrime(num),
		IsPerfect:   isPerfect(num),
		Properties:  Properties(num),
		SumOfDigits: sumOfDigits(num),
		FunFact:     fetchFunFact(num),
		Error:       false,
	}

	// Send valid JSON response
	json.NewEncoder(w).Encode(response)
}

// Run initializes the router
func Run() {
	r := mux.NewRouter()
	r.HandleFunc("/api/classify-number", get).Methods("GET")

	log.Fatal(http.ListenAndServe(":8000", r))
}

func main() {
	Run()
}
