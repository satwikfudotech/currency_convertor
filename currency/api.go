package currency

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const apiURL = "https://api.apilayer.com/fixer/latest?base=%s&symbols=%s"

var (
	APIKey      = "a18eedf762b833a0b7bc7c05ac5515bf"
	ratesCache  = NewCache()
	cacheTicker *time.Ticker
)

type ApiResponse struct {
	Rates map[string]float64 `json:"rates"`
}

func Init(apiKey string) {
	APIKey = apiKey
	cacheTicker = time.NewTicker(30 * time.Minute)
	go func() {
		for range cacheTicker.C {
			ratesCache.Clear()
		}
	}()
}

func GetExchangeRate(base, target string) (float64, error) {
	if rate, ok := ratesCache.Get(target); ok {
		return rate, nil
	}

	client := &http.Client{}
	req, _ := http.NewRequest("GET", fmt.Sprintf(apiURL, base, target), nil)
	req.Header.Set("apikey", APIKey)

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, errors.New("API request failed")
	}

	var data ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return 0, err
	}

	rate, exists := data.Rates[target]
	if !exists {
		return 0, errors.New("currency not found")
	}

	ratesCache.Set(target, rate)
	return rate, nil
}
