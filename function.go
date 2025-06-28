package function

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type RateRequest struct {
	Base string `json:"base"`
}

type RateResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

func Converter(w http.ResponseWriter, r *http.Request) {
	var req RateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Base == "" {
		http.Error(w, "Base currency is required", http.StatusBadRequest)
		return
	}

	fmt.Printf("Received base: %s\n", req.Base)

	url := fmt.Sprintf("https://api.exchangerate.host/latest?base=%s", req.Base)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("HTTP GET failed: %v\n", err)
		http.Error(w, "Failed to fetch exchange rates", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("Exchange API returned status: %s\n", resp.Status)
		http.Error(w, "Exchange API error", http.StatusInternalServerError)
		return
	}

	var rateResp RateResponse
	if err := json.NewDecoder(resp.Body).Decode(&rateResp); err != nil {
		fmt.Printf("Failed to parse JSON: %v\n", err)
		http.Error(w, "Failed to parse exchange data", http.StatusInternalServerError)
		return
	}

	if rateResp.Base == "" || rateResp.Rates == nil {
		http.Error(w, "Exchange API returned empty data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rateResp)
}
