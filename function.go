package function

import (
	"encoding/json"
	"net/http"
	"os"
	
	"example.com/currency-converter/currency"
)

type Request struct {
	Amount  float64 `json:"amount"`
	From    string  `json:"from"`
	To      string  `json:"to"`
	UsePtrs bool    `json:"use_pointers"`
}

type Response struct {
	Converted float64 `json:"converted"`
	Currency  string  `json:"currency"`
	Method    string  `json:"method"`
}

func init() {
	apiKey := os.Getenv("FIXER_API_KEY")
	if apiKey == "" {
		panic("FIXER_API_KEY environment variable not set")
	}
	currency.Init(apiKey)
}

// Converter is the HTTP handler for Google Cloud Functions
func Converter(w http.ResponseWriter, r *http.Request) {
	// Handle CORS preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Set CORS headers for the main request
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	fromRate, err := currency.GetExchangeRate("USD", req.From)
	if err != nil {
		http.Error(w, `{"error":"Error getting source currency rate"}`, http.StatusInternalServerError)
		return
	}

	toRate, err := currency.GetExchangeRate("USD", req.To)
	if err != nil {
		http.Error(w, `{"error":"Error getting target currency rate"}`, http.StatusInternalServerError)
		return
	}

	var result float64
	method := "Without Pointers"
	if req.UsePtrs {
		result = currency.ConvertWithPointers(req.Amount, &fromRate, &toRate)
		method = "With Pointers"
	} else {
		result = currency.ConvertWithoutPointers(req.Amount, fromRate, toRate)
	}

	json.NewEncoder(w).Encode(Response{
		Converted: result,
		Currency:  req.To,
		Method:    method,
	})
}
