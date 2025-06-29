package function

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Input request structure
type RateRequest struct {
	Base string `json:"base"` // Example: "INR", "USD", etc.
}

// Output response structure
type RateResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

// Main function to handle HTTP request
func Converter(w http.ResponseWriter, r *http.Request) {
	// Parse JSON request body
	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// If base currency is not provided
	if req.Base == "" {
		http.Error(w, "Base currency is required", http.StatusBadRequest)
		return
	}

	// Build URL using free ExchangeRate.host API
	url := fmt.Sprintf("https://api.exchangerate.host/latest?base=%s", req.Base)

	// Make the HTTP request to fetch exchange rates
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Failed to fetch exchange rates", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse API response into RateResponse struct
	var rateResp RateResponse
	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		http.Error(w, "Failed to parse exchange data", http.StatusInternalServerError)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rateResp)
}
