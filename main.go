package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-proxy/exchanges"

	"github.com/joho/godotenv"
)

// Define the structure for the API response
type Proxy struct {
	ProxyAddress string `json:"proxy_address"`
	Port         int    `json:"port"`
	CountryCode  string `json:"country_code"`
}

type ApiResponse struct {
	Next    string  `json:"next"`
	Results []Proxy `json:"results"`
}

// Cache file path
const cacheFile = "proxy_cache.json"

// Helper function to format table with proper column alignment
func formatTableRow(exchange, proxyAddr string, port int, country, responseTime, data string) string {
	return fmt.Sprintf("%-12s %-15s %-6d %-8s %-15s %s", exchange, proxyAddr, port, country, responseTime, data)
}

// Helper function to calculate response time statistics
func calculateResponseTimeStats(results []*exchanges.TestResult) (min, max, avg, median time.Duration) {
	if len(results) == 0 {
		return 0, 0, 0, 0
	}

	// Extract response times
	var times []time.Duration
	for _, result := range results {
		times = append(times, result.ResponseTime)
	}

	// Calculate min and max
	min = times[0]
	max = times[0]
	var total time.Duration
	for _, t := range times {
		if t < min {
			min = t
		}
		if t > max {
			max = t
		}
		total += t
	}

	// Calculate average
	avg = total / time.Duration(len(times))

	// Calculate median
	sort.Slice(times, func(i, j int) bool {
		return times[i] < times[j]
	})
	if len(times)%2 == 0 {
		median = (times[len(times)/2-1] + times[len(times)/2]) / 2
	} else {
		median = times[len(times)/2]
	}

	return min, max, avg, median
}

func handleListCommand() {
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

func handleApiCommand() {
	// Check for --refresh flag
	refresh := false
	if len(os.Args) > 2 && os.Args[2] == "--refresh" {
		refresh = true
	}

	// If not refreshing, try to load from cache first
	if !refresh {
		if cachedProxies, err := loadFromCache(); err == nil {
			fmt.Println("Proxy Address\tPort\tCountry")
			for _, proxy := range cachedProxies {
				fmt.Printf("%s\t%d\t%s\n", proxy.ProxyAddress, proxy.Port, proxy.CountryCode)
			}
			fmt.Fprintf(os.Stderr, "\nTotal proxies loaded from cache: %d\n", len(cachedProxies))
			return
		}
	}

	// Get the API key from environment variable
	apiKey := os.Getenv("PROXY_API")
	if apiKey == "" {
		fmt.Println("PROXY_API environment variable is not set")
		return
	}

	baseURL := "https://proxy.webshare.io/api/v2/proxy/list/?mode=direct&page_size=100"
	url := baseURL

	client := &http.Client{}
	fmt.Println("Proxy Address\tPort\tCountry")

	var allProxies []Proxy
	totalProxies := 0
	pageCount := 0

	for url != "" {
		pageCount++
		fmt.Fprintf(os.Stderr, "Fetching page %d...\n", pageCount)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatalf("Error creating request: %v", err)
		}
		req.Header.Set("Authorization", "Token "+apiKey)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatalf("Error making request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Fatalf("API request failed with status: %s", resp.Status)
		}

		var apiResp ApiResponse
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&apiResp); err != nil {
			log.Fatalf("Error decoding JSON: %v", err)
		}

		for _, proxy := range apiResp.Results {
			fmt.Printf("%s\t%d\t%s\n", proxy.ProxyAddress, proxy.Port, proxy.CountryCode)
			allProxies = append(allProxies, proxy)
			totalProxies++
		}

		url = apiResp.Next // Move to next page (or exit loop if empty)
	}

	// Save to cache
	if err := saveToCache(allProxies); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Failed to save to cache: %v\n", err)
	}

	fmt.Fprintf(os.Stderr, "\nTotal proxies fetched: %d (across %d pages)\n", totalProxies, pageCount)
}

func loadFromCache() ([]Proxy, error) {
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var proxies []Proxy
	if err := json.Unmarshal(data, &proxies); err != nil {
		return nil, err
	}

	return proxies, nil
}

func saveToCache(proxies []Proxy) error {
	data, err := json.MarshalIndent(proxies, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
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

func handleTestCommand() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./go-proxy test <exchange> [options]")
		fmt.Println("Available exchanges:")
		registry := exchanges.NewRegistry()
		for _, name := range registry.List() {
			fmt.Printf("  %s\n", name)
		}
		fmt.Println("Use '*' to test all available exchanges")
		fmt.Println("Options:")
		fmt.Println("  --limit <number> - Limit the number of proxies to test (e.g., --limit 10)")
		return
	}

	exchangeName := os.Args[2]

	// Parse command line flags
	limit := -1 // -1 means no limit
	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "--limit" && i+1 < len(os.Args) {
			if val, err := strconv.Atoi(os.Args[i+1]); err == nil && val > 0 {
				limit = val
			} else {
				fmt.Printf("Error: Invalid limit value '%s'. Must be a positive integer.\n", os.Args[i+1])
				return
			}
			// Remove the --limit and its value from os.Args to avoid confusion
			os.Args = append(os.Args[:i], os.Args[i+2:]...)
			break
		}
	}

	proxies, err := loadFromCache()
	if err != nil {
		fmt.Printf("Error loading proxies from cache: %v\n", err)
		fmt.Println("Please run './go-proxy api' first to fetch proxies")
		return
	}
	if len(proxies) == 0 {
		fmt.Println("No proxies found in cache. Please run './go-proxy api' first to fetch proxies")
		return
	}

	// Apply limit if specified
	if limit > 0 && limit < len(proxies) {
		proxies = proxies[:limit]
		fmt.Printf("Limited to first %d proxies from cache\n", limit)
	}

	fmt.Printf("Testing %d proxies...\n", len(proxies))

	var testers []exchanges.ExchangeTester
	registry := exchanges.NewRegistry()
	if exchangeName == "*" {
		for _, name := range registry.List() {
			tester, err := registry.Get(name)
			if err != nil {
				fmt.Printf("Warning: Could not get tester for %s: %v\n", name, err)
				continue
			}
			testers = append(testers, tester)
		}
	} else {
		tester, err := registry.Get(exchangeName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("Available exchanges:")
			for _, name := range registry.List() {
				fmt.Printf("  %s\n", name)
			}
			return
		}
		testers = append(testers, tester)
	}

	// Get concurrency limit from env or default to 10
	concurrency := 10
	if val := os.Getenv("PROXY_TEST_CONCURRENCY"); val != "" {
		if n, err := strconv.Atoi(val); err == nil && n > 0 {
			concurrency = n
		}
	}

	var wg sync.WaitGroup
	results := make(chan *exchanges.TestResult, len(proxies)*len(testers))
	semaphore := make(chan struct{}, concurrency)
	totalTests := len(proxies) * len(testers)

	// Progress tracking
	var progressMutex sync.Mutex
	completedTests := 0

	for _, tester := range testers {
		exchangeName := tester.GetName()
		for _, proxy := range proxies {
			wg.Add(1)
			go func(tester exchanges.ExchangeTester, proxy Proxy, exchangeName string) {
				defer wg.Done()
				semaphore <- struct{}{}
				defer func() { <-semaphore }()
				// Optionally: Add retry logic here (simple 1 retry for transient errors)
				var result *exchanges.TestResult
				var err error
				for attempt := 1; attempt <= 2; attempt++ {
					result, err = tester.TestProxy(proxy.ProxyAddress, proxy.Port)
					if err == nil && result.Success {
						break
					}
					time.Sleep(500 * time.Millisecond)
				}
				if result == nil {
					result = &exchanges.TestResult{
						ProxyAddress: proxy.ProxyAddress,
						Port:         proxy.Port,
						CountryCode:  proxy.CountryCode,
						Success:      false,
						Error:        fmt.Sprintf("Test error: %v", err),
						ResponseTime: 0,
					}
				}
				result.Exchange = exchangeName
				result.CountryCode = proxy.CountryCode
				results <- result

				// Update progress and print result immediately
				progressMutex.Lock()
				completedTests++
				currentProgress := completedTests
				progressMutex.Unlock()

				// Print each test result as it completes
				if result.Success {
					fmt.Printf("[%3d/%3d] ✅ %-12s - %-15s:%-5d (%-2s) - %-12s - %s\n",
						currentProgress, totalTests,
						result.Exchange,
						result.ProxyAddress,
						result.Port,
						result.CountryCode,
						result.ResponseTime.String(),
						result.Data)
				} else {
					fmt.Printf("[%3d/%3d] ❌ %-12s - %-15s:%-5d (%-2s) - %s\n",
						currentProgress, totalTests,
						result.Exchange,
						result.ProxyAddress,
						result.Port,
						result.CountryCode,
						result.Error)
				}
			}(tester, proxy, exchangeName)
		}
	}

	go func() {
		wg.Wait()
		close(results)
		fmt.Printf("\n=== All tests completed ===\n")
	}()

	var successfulTests []*exchanges.TestResult
	var failedTests []*exchanges.TestResult
	for result := range results {
		if result.Success {
			successfulTests = append(successfulTests, result)
		} else {
			failedTests = append(failedTests, result)
		}
	}

	fmt.Printf("\n=== Test Results ===\n")
	fmt.Printf("Successful tests: %d\n", len(successfulTests))
	fmt.Printf("Failed tests: %d\n", len(failedTests))
	fmt.Printf("Total tests: %d\n", len(successfulTests)+len(failedTests))

	if len(successfulTests) > 0 {
		fmt.Printf("\n=== Successful Tests ===\n")
		fmt.Printf("%-12s %-15s %-6s %-8s %-15s %s\n", "Exchange", "Proxy Address", "Port", "Country", "Response Time", "Data")
		fmt.Println(strings.Repeat("-", 90)) // Separator line
		for _, result := range successfulTests {
			fmt.Println(formatTableRow(
				result.Exchange,
				result.ProxyAddress,
				result.Port,
				result.CountryCode,
				result.ResponseTime.String(),
				result.Data,
			))
		}

		// Calculate and display response time statistics
		min, max, avg, median := calculateResponseTimeStats(successfulTests)
		fmt.Printf("\nResponse Time Statistics:\n")
		fmt.Printf("Min: %s\n", min.String())
		fmt.Printf("Max: %s\n", max.String())
		fmt.Printf("Avg: %s\n", avg.String())
		fmt.Printf("Median: %s\n", median.String())
	}

	if len(failedTests) > 0 {
		fmt.Printf("\n=== Failed Tests ===\n")
		fmt.Printf("%-12s %-15s %-6s %-8s %s\n", "Exchange", "Proxy Address", "Port", "Country", "Error")
		fmt.Println(strings.Repeat("-", 70)) // Separator line
		for _, result := range failedTests {
			fmt.Printf("%-12s %-15s %-6d %-8s %s\n",
				result.Exchange,
				result.ProxyAddress,
				result.Port,
				result.CountryCode,
				result.Error)
		}
	}
}

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		fmt.Printf("Error loading .env file: %v\n", err)
		return
	}

	// Check command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./go-proxy <command> [options]")
		fmt.Println("Commands:")
		fmt.Println("  list - Download and display proxy list from PROXY_LIST URL")
		fmt.Println("  api  - Fetch proxy list from API using PROXY_API key")
		fmt.Println("  test - Test proxies with exchange APIs")
		fmt.Println("Options for api command:")
		fmt.Println("  --refresh - Force refresh the cached proxy list")
		fmt.Println("Options for test command:")
		fmt.Println("  <exchange> - Specific exchange to test (e.g., binance)")
		fmt.Println("  * - Test all available exchanges")
		fmt.Println("  --limit <number> - Limit the number of proxies to test (e.g., --limit 10)")
		return
	}

	command := os.Args[1]

	switch command {
	case "list":
		handleListCommand()
	case "api":
		handleApiCommand()
	case "test":
		handleTestCommand()
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Available commands: list, api, test")
	}
}
