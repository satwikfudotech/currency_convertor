package function

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RateRequest struct {
	Base string `json:"base"` // "INR", "USD", etc.
}

type RateResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func Converter(w http.ResponseWriter, r *http.Request) {
	// Parse incoming JSON
	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Log incoming base currency
	fmt.Println("Received base:", req.Base)

	if req.Base == "" {
		http.Error(w, "Base currency is required", http.StatusBadRequest)
		return
	}

	// Build API request
	url := fmt.Sprintf("https://api.exchangerate.host/latest?base=%s", req.Base)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Failed to fetch rates", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Parse API response
	var rateResp RateResponse
	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		http.Error(w, "Error decoding API response", http.StatusInternalServerError)
		return
	}

	// Handle empty API response
	if rateResp.Base == "" || rateResp.Rates == nil {
		http.Error(w, "Invalid currency code or empty API response", http.StatusBadRequest)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rateResp)
}
