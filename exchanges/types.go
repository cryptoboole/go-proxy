package exchanges

import (
	"fmt"
	"net/url"
	"os"
	"time"
)

// TestResult represents the result of testing a proxy with an exchange
type TestResult struct {
	Exchange     string        `json:"exchange"`
	ProxyAddress string        `json:"proxy_address"`
	Port         int           `json:"port"`
	CountryCode  string        `json:"country_code,omitempty"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
	ResponseTime time.Duration `json:"response_time"`
	Data         string        `json:"data,omitempty"`
}

// ExchangeTester interface defines methods that all exchange testers must implement
type ExchangeTester interface {
	TestProxy(proxyAddress string, port int) (*TestResult, error)
	GetName() string
}

// CreateProxyURL creates a proper URL for proxy configuration with authentication
func CreateProxyURL(proxyAddress string, port int) (*url.URL, error) {
	username := os.Getenv("PROXY_USER")
	password := os.Getenv("PROXY_PASS")

	if username != "" && password != "" {
		// Include authentication in the proxy URL
		return url.Parse(fmt.Sprintf("http://%s:%s@%s:%d", username, password, proxyAddress, port))
	}

	// No authentication
	return url.Parse(fmt.Sprintf("http://%s:%d", proxyAddress, port))
}
