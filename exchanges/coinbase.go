package exchanges

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// CoinbaseResponse represents the structure of Coinbase API response
type CoinbaseResponse struct {
	Data struct {
		Base     string `json:"base"`
		Currency string `json:"currency"`
		Amount   string `json:"amount"`
	} `json:"data"`
}

// CoinbaseTester implements the ExchangeTester interface for Coinbase
type CoinbaseTester struct {
	client *http.Client
}

// NewCoinbaseTester creates a new Coinbase tester instance
func NewCoinbaseTester() *CoinbaseTester {
	return &CoinbaseTester{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TestProxy tests if a proxy works with Coinbase API
func (c *CoinbaseTester) TestProxy(proxyAddress string, port int) (*TestResult, error) {
	// Create proxy URL
	proxyURL, err := CreateProxyURL(proxyAddress, port)
	if err != nil {
		return &TestResult{
			ProxyAddress: proxyAddress,
			Port:         port,
			Success:      false,
			Error:        fmt.Sprintf("Invalid proxy URL: %v", err),
			ResponseTime: 0,
		}, nil
	}

	// Create transport with proxy
	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	// Create client with custom transport
	client := &http.Client{
		Transport: transport,
		Timeout:   10 * time.Second,
	}

	// Test endpoint - Coinbase spot price for BTC-USD
	testURL := "https://api.coinbase.com/v2/prices/BTC-USD/spot"

	startTime := time.Now()

	req, err := http.NewRequest("GET", testURL, nil)
	if err != nil {
		return &TestResult{
			ProxyAddress: proxyAddress,
			Port:         port,
			Success:      false,
			Error:        fmt.Sprintf("Failed to create request: %v", err),
			ResponseTime: time.Since(startTime),
		}, nil
	}

	resp, err := client.Do(req)
	if err != nil {
		return &TestResult{
			ProxyAddress: proxyAddress,
			Port:         port,
			Success:      false,
			Error:        fmt.Sprintf("Request failed: %v", err),
			ResponseTime: time.Since(startTime),
		}, nil
	}
	defer resp.Body.Close()

	responseTime := time.Since(startTime)

	// Check if response is successful
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return &TestResult{
			ProxyAddress: proxyAddress,
			Port:         port,
			Success:      false,
			Error:        fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(body)),
			ResponseTime: responseTime,
		}, nil
	}

	// Try to parse the response to ensure it's valid JSON
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &TestResult{
			ProxyAddress: proxyAddress,
			Port:         port,
			Success:      false,
			Error:        fmt.Sprintf("Failed to read response: %v", err),
			ResponseTime: responseTime,
		}, nil
	}

	var coinbaseResp CoinbaseResponse
	if err := json.Unmarshal(body, &coinbaseResp); err != nil {
		return &TestResult{
			ProxyAddress: proxyAddress,
			Port:         port,
			Success:      false,
			Error:        fmt.Sprintf("Invalid JSON response: %v", err),
			ResponseTime: responseTime,
		}, nil
	}

	// Verify we got expected data
	if coinbaseResp.Data.Base != "BTC" || coinbaseResp.Data.Currency != "USD" || coinbaseResp.Data.Amount == "" {
		return &TestResult{
			ProxyAddress: proxyAddress,
			Port:         port,
			Success:      false,
			Error:        "Unexpected response format",
			ResponseTime: responseTime,
		}, nil
	}

	return &TestResult{
		ProxyAddress: proxyAddress,
		Port:         port,
		Success:      true,
		ResponseTime: responseTime,
		Data:         fmt.Sprintf("BTC Price: $%s", coinbaseResp.Data.Amount),
	}, nil
}

// GetName returns the exchange name
func (c *CoinbaseTester) GetName() string {
	return "Coinbase"
}
