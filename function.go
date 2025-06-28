package function

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
)

// Request format: accepts amount, source currency, and multiple target currencies
type MultiConvertRequest struct {
	Amount float64  `json:"amount"`
	From   string   `json:"from"`
	To     []string `json:"to"` // multiple targets
}

// Response format: returns converted values for each target currency
type MultiConvertResponse struct {
	Results map[string]float64 `json:"results"`
}

// Fixer API response structure
type FixerAPIResponse struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
}

// Google Cloud Function entry point
func Converter(w http.ResponseWriter, r *http.Request) {
	var req MultiConvertRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON input", http.StatusBadRequest)
		return
	}

	apiKey := os.Getenv("FIXER_API_KEY")
	if apiKey == "" {
		http.Error(w, "FIXER_API_KEY not found in env", http.StatusInternalServerError)
		return
	}

	// Call Fixer.io API
	url := fmt.Sprintf("https://data.fixer.io/api/latest?access_key=%s", apiKey)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		http.Error(w, "Failed to fetch exchange rates", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var fixerData FixerAPIResponse
	json.Unmarshal(body, &fixerData)

	fromRate := fixerData.Rates[req.From]
	if fromRate == 0 {
		http.Error(w, "Source currency not supported", http.StatusBadRequest)
		return
	}

	// Parallel conversion using goroutines and WaitGroup
	results := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, toCurrency := range req.To {
		wg.Add(1)
		go func(to string) {
			defer wg.Done()
			toRate := fixerData.Rates[to]
			if toRate == 0 {
				return // skip unsupported
			}
			eur := req.Amount / fromRate
			converted := eur * toRate

			mu.Lock()
			results[to] = converted
			mu.Unlock()
		}(toCurrency)
	}

	wg.Wait()

	// Return all converted results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(MultiConvertResponse{Results: results})
}
