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

func Converter(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "OPTIONS" {
		w.Header().Set("Access-Control-Allow-Methods", "POST")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(204)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendError(w, "Invalid JSON format", 400)
		return
	}

	if req.Amount <= 0 {
		sendError(w, "Amount must be positive", 400)
		return
	}

	req.From = strings.ToUpper(strings.TrimSpace(req.From))
	req.To = strings.ToUpper(strings.TrimSpace(req.To))
	if len(req.From) != 3 || len(req.To) != 3 {
		sendError(w, "Currency codes must be 3 letters", 400)
		return
	}

	apiKey := os.Getenv("FIXER_API_KEY")
	if apiKey == "" {
		sendError(w, "Fixer API key not configured", 500)
		return
	}

	rates, err := getRates(apiKey, req.From, req.To)
	if err != nil {
		sendError(w, fmt.Sprintf("API error: %v", err), 500)
		return
	}

	fromRate, ok := rates[req.From]
	if !ok {
		sendError(w, "Invalid source currency", 400)
		return
	}

	toRate, ok := rates[req.To]
	if !ok {
		sendError(w, "Invalid target currency", 400)
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

	json.NewEncoder(w).Encode(Response{
		Converted: converted,
		Currency:  req.To,
		Method:    method,
	})
}

func getRates(apiKey, from, to string) (map[string]float64, error) {
	url := fmt.Sprintf("https://api.apilayer.com/fixer/latest?symbols=%s,%s", from, to)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("apikey", apiKey)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("Fixer API error [%d]: %s", res.StatusCode, string(body))
	}

	var result struct {
		Rates map[string]float64 `json:"rates"`
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("invalid API response: %v", err)
	}

	return result.Rates, nil
}

func sendError(w http.ResponseWriter, message string, code int) {
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(Response{Error: message})
}

func convertWithoutPointers(amount, fromRate, toRate float64) float64 {
	return amount * (toRate / fromRate)
}

func convertWithPointers(amount float64, fromRate, toRate *float64) float64 {
	return amount * (*toRate / *fromRate)
}
