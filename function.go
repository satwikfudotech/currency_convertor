package function

import (
	
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Structs for request and response
type ConvertRequest struct {
	Amount float64 `json:"amount"`
	From   string  `json:"from"`
	To     string  `json:"to"`
}

type ConvertResponse struct {
	Converted float64 `json:"converted"`
	Message   string  `json:"message"`
}

// Struct for Fixer API response
type FixerAPIResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

// Main entry point for Cloud Function
func ConvertCurrency(w http.ResponseWriter, r *http.Request) {
	var req ConvertRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("FIXER_API_KEY")
	if apiKey == "" {
		http.Error(w, "Fixer API key not found", http.StatusInternalServerError)
		return
	}

	// Fetch exchange rates
	url := fmt.Sprintf("https://data.fixer.io/api/latest?access_key=%s", apiKey)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Failed to fetch rates", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var fixerData FixerAPIResponse
	json.Unmarshal(body, &fixerData)

	fromRate := fixerData.Rates[req.From]
	toRate := fixerData.Rates[req.To]

	if fromRate == 0 || toRate == 0 {
		http.Error(w, "Currency not supported", http.StatusBadRequest)
		return
	}

	// Convert: amount -> EUR -> target
	eurAmount := req.Amount / fromRate
	converted := eurAmount * toRate

	// Send response
	json.NewEncoder(w).Encode(ConvertResponse{
		Converted: converted,
		Message:   fmt.Sprintf("%.2f %s = %.2f %s", req.Amount, req.From, converted, req.To),
	})
}

// For Cloud Functions deployment
func init() {
	// Use the Functions Framework routing
	http.HandleFunc("/", ConvertCurrency)
}
