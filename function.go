package function

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Request format
type RateRequest struct {
	Base string `json:"base"` // Example: "INR", "USD", etc.
}

// Response format
type RateResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

// Handler function for Cloud Function
func Converter(w http.ResponseWriter, r *http.Request) {
	// Decode the JSON body into RateRequest struct
	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Base == "" {
		http.Error(w, "Base currency is required", http.StatusBadRequest)
		return
	}

	// Build URL for ExchangeRate.host API
	url := fmt.Sprintf("https://api.exchangerate.host/latest?base=%s", req.Base)

	// Make the HTTP GET request
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Failed to fetch exchange rates", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse response into RateResponse struct
	var rateResp RateResponse
	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		http.Error(w, "Failed to parse exchange rate data", http.StatusInternalServerError)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rateResp)
}
