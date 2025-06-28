package currency

import "sync"

type RateCache struct {
	rates map[string]float64
	mu    sync.RWMutex
}

func NewCache() *RateCache {
	return &RateCache{
		rates: make(map[string]float64),
	}
}

func (c *RateCache) Get(currency string) (float64, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rate, exists := c.rates[currency]
	return rate, exists
}

func (c *RateCache) Set(currency string, rate float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rates[currency] = rate
}

func (c *RateCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.rates = make(map[string]float64)
}
