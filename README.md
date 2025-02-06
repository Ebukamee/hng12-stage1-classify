# Number Classification API

This is a REST API built using Go and Gorilla Mux that classifies numbers based on their mathematical properties. The API determines whether a number is **prime**, **perfect**, **Armstrong**, **even/odd**, calculates the **sum of its digits**, and fetches an interesting mathematical fact about it.

## Features
- Checks if a number is **prime**
- Checks if a number is **perfect**
- Checks if a number is **Armstrong**
- Determines if a number is **even or odd**
- Calculates the **sum of digits**
- Fetches a **fun mathematical fact** about the number from NumbersAPI
- Handles error cases for **invalid input types**

## Technologies Used
- **Go**: The programming language used for development
- **Gorilla Mux**: A powerful HTTP router and dispatcher for Go
- **JSON Encoding**: Used to format the response in JSON
- **NumbersAPI**: Used to fetch fun facts about numbers

## Installation & Setup
### Prerequisites
Ensure you have Go installed on your system. You can download it from [Go's official website](https://go.dev/).

### Clone the Repository
```sh
git clone https://github.com/yourusername/number-classification-api.git
cd number-classification-api
```

### Install Dependencies
```sh
go mod tidy
```

### Run the Application
```sh
go run main.go
```

The server will start on port `8000`. You can access it by visiting:
```
http://localhost:8000/
```

Once hosted, use:
```
https://hng12-stage1-classify.onrender.com/api/classify-number
```

## API Endpoint
### `GET /api/classify-number?number={number}`
#### Parameters:
- `number`: The number to classify (must be an integer).

#### Example Request:
```
GET /api/classify-number?number=6
```

#### Example Response:
```json
{
  "number": 6,
  "is_prime": false,
  "is_perfect": true,
  "properties": ["even"],
  "digit_sum": 6,
  "fun_fact": "6 is the smallest perfect number.",
  "error": false
}
```

## Error Handling
The API returns an error response if the input is invalid:

| Error Type  | Description | Example Response |
|------------|------------|------------------|
| **Null**   | If no number is provided | `{ "number": "null", "error": true }` |
| **Alphabet** | If input contains only letters | `{ "number": "alphabet", "error": true }` |
| **Invalid** | If input is not a valid number | `{ "number": "invalid", "error": true }` |

## Project Structure
```
.
├── go.mod
├── go.sum
├── main.go
└── README.md
```

## License
This project is open-source and available under the [MIT License](LICENSE).

## Author
**Ebuka**

## Backlink
Looking to Hire a Go developer? Check out [HNG's talent pool](https://hng.tech/hire/golang-developers).

