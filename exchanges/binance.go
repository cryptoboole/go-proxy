package exchanges

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// BinanceResponse represents the structure of Binance API response
type BinanceResponse struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// BinanceTester implements the ExchangeTester interface for Binance
type BinanceTester struct {
	client *http.Client
}

// NewBinanceTester creates a new Binance tester instance
func NewBinanceTester() *BinanceTester {
	return &BinanceTester{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// TestProxy tests if a proxy works with Binance API
func (b *BinanceTester) TestProxy(proxyAddress string, port int) (*TestResult, error) {
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

	// Test endpoint - Binance ticker price for BTC/USDT
	testURL := "https://api.binance.com/api/v3/ticker/price?symbol=BTCUSDT"

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

	var binanceResp BinanceResponse
	if err := json.Unmarshal(body, &binanceResp); err != nil {
		return &TestResult{
			ProxyAddress: proxyAddress,
			Port:         port,
			Success:      false,
			Error:        fmt.Sprintf("Invalid JSON response: %v", err),
			ResponseTime: responseTime,
		}, nil
	}

	// Verify we got expected data
	if binanceResp.Symbol != "BTCUSDT" || binanceResp.Price == "" {
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
		Data:         fmt.Sprintf("BTC Price: %s", binanceResp.Price),
	}, nil
}

// GetName returns the exchange name
func (b *BinanceTester) GetName() string {
	return "Binance"
}
