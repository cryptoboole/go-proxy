# Go Proxy CLI

A command-line tool for managing and testing proxy lists with cryptocurrency exchange APIs.

## Usage

```bash
./go-proxy <command> [options]
```

### Commands

- `list` - Download and display proxy list from PROXY_LIST URL
- `api` - Fetch proxy list from API using PROXY_API key (with caching)
- `test` - Test proxies with cryptocurrency exchange APIs

### Options

**For `api` command:**
- `--refresh` - Force refresh the cached proxy list

**For `test` command:**
- `<exchange>` - Specific exchange to test (e.g., `binance`, `coinbase`)
- `*` - Test all available exchanges
- `--limit <number>` - Limit the number of proxies to test (e.g., `--limit 10`)

## Environment Variables

Create a `.env` file in the project root with the following variables:

```env
# Proxy list URL for the 'list' command
PROXY_LIST=https://example.com/proxy-list.txt

# API key for the 'api' command (Webshare.io API)
PROXY_API=your_api_key_here

# Optional: Proxy authentication
PROXY_USER=your_proxy_username
PROXY_PASS=your_proxy_password

# Optional: Concurrency limit for testing (default: 10)
PROXY_TEST_CONCURRENCY=10
```

## Examples

```bash
# List proxies from URL
./go-proxy list

# Fetch proxies from API (cached)
./go-proxy api

# Force refresh cached proxies
./go-proxy api --refresh

# Test all proxies with Binance API
./go-proxy test binance

# Test all proxies with all available exchanges
./go-proxy test *

# Test first 5 proxies with Coinbase API
./go-proxy test coinbase --limit 5
```

## Features

- **Proxy Management**: Download from URLs or fetch from APIs
- **Caching**: Proxy lists are cached to avoid repeated API calls
- **Exchange Testing**: Test proxies against cryptocurrency exchanges
- **Concurrent Testing**: Multiple proxies tested simultaneously for efficiency
- **Detailed Results**: Response times, success/failure rates, and error reporting
- **Statistics**: Min, max, average, and median response time calculations

## Supported Exchanges

- **Binance**: Tests against Binance API endpoints
- **Coinbase**: Tests against Coinbase API endpoints

## Building

```bash
go build -o go-proxy main.go
``` 