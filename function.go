package function

import (
	"encoding/json"
	"fmt"
	"io"
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

	fmt.Println("Received base:", req.Base)

	if req.Base == "" {
		http.Error(w, "Base currency is required", http.StatusBadRequest)
		return
	}

	// ✅ Free API, no key required
	url := fmt.Sprintf("https://api.exchangerate.host/latest?base=%s", req.Base)

	resp, err := http.Get(url)
	if err != nil {
		http.Error(w, "Failed to fetch exchange rates", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Println("API Raw Response:", string(bodyBytes))

	var rateResp RateResponse
	if err := json.Unmarshal(bodyBytes, &rateResp); err != nil {
		http.Error(w, "Failed to parse API response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rateResp)
}
