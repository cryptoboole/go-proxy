package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	// Get the proxy list URL from environment variable
	proxyListURL := os.Getenv("PROXY_LIST")
	if proxyListURL == "" {
		fmt.Println("PROXY_LIST environment variable is not set")
		return
	}

	// Download the proxy list
	fmt.Printf("Downloading proxy list from: %s\n", proxyListURL)
	proxies, err := downloadProxyList(proxyListURL)
	if err != nil {
		fmt.Printf("Error downloading proxy list: %v\n", err)
		return
	}

	// Print out the proxies
	fmt.Printf("Found %d proxies:\n", len(proxies))
	for i, proxy := range proxies {
		fmt.Printf("%d. %s\n", i+1, strings.TrimSpace(proxy))
	}
}

func downloadProxyList(url string) ([]string, error) {
	// Make HTTP request to download the proxy list
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to download proxy list: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status: %d", resp.StatusCode)
	}

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// Split the content by lines and filter out empty lines
	lines := strings.Split(string(body), "\n")
	var proxies []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			proxies = append(proxies, line)
		}
	}

	return proxies, nil
}
