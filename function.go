package function

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type Request struct {
	Amount     float64 `json:"amount"`
	From       string  `json:"from"`
	To         string  `json:"to"`
	UsePointers bool   `json:"use_pointers"`
}

type Response struct {
	Converted float64 `json:"converted"`
	Currency  string  `json:"currency"`
	Method    string  `json:"method"`
	Error     string  `json:"error,omitempty"`
}

// Handler for Google Cloud Function
func Converter(w http.ResponseWriter, r *http.Request) {
	// Allow CORS for frontend usage
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodOptions {
		// Handle preflight request
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request body"}`, http.StatusBadRequest)
		return
	}

	// Validate input
	if req.Amount <= 0 {
		http.Error(w, `{"error":"Amount must be positive"}`, http.StatusBadRequest)
		return
	}
	if len(req.From) != 3 || len(req.To) != 3 {
		http.Error(w, `{"error":"Currency codes must be 3 letters"}`, http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("FIXER_API_KEY")
	if apiKey == "" {
		http.Error(w, `{"error":"API key not set"}`, http.StatusInternalServerError)
		return
	}

	from := strings.ToUpper(req.From)
	to := strings.ToUpper(req.To)

	// Get rates from Fixer.io
	rates, err := getRates(apiKey, from, to)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%v"}`, err), http.StatusInternalServerError)
		return
	}

	fromRate := 1.0
	if from != "EUR" {
		r, ok := rates[from]
		if !ok {
			http.Error(w, `{"error":"Invalid source currency"}`, http.StatusBadRequest)
			return
		}
		fromRate = r
	}
	toRate, ok := rates[to]
	if !ok {
		http.Error(w, `{"error":"Invalid target currency"}`, http.StatusBadRequest)
		return
	}

	var converted float64
	method := "Without Pointers"
	if req.UsePointers {
		converted = convertWithPointers(req.Amount, &fromRate, &toRate)
		method = "With Pointers"
	} else {
		converted = convertWithoutPointers(req.Amount, fromRate, toRate)
	}

	resp := Response{
		Converted: converted,
		Currency:  to,
		Method:    method,
	}
	json.NewEncoder(w).Encode(resp)
}

// Fetch rates from Fixer.io
func getRates(apiKey, from, to string) (map[string]float64, error) {
	// Fixer.io free plan only allows EUR as base, so we fetch both rates relative to EUR
	url := fmt.Sprintf("https://api.apilayer.com/fixer/latest?symbols=%s,%s", from, to)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Fixer.io: %v", err)
	}
	defer res.Body.Close()

	body, _ := io.ReadAll(res.Body)
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("Fixer.io error: %s", body)
	}

	var data struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("failed to parse Fixer.io response")
	}
	return data.Rates, nil
}

// Conversion functions
func convertWithoutPointers(amount, fromRate, toRate float64) float64 {
	return amount * (toRate / fromRate)
}
func convertWithPointers(amount float64, fromRate, toRate *float64) float64 {
	return amount * (*toRate / *fromRate)
}
