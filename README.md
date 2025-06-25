# Go Proxy CLI

A command-line tool for managing proxy lists with two main commands.

## Usage

```bash
./go-proxy <command>
```

### Commands

- `list` - Download and display proxy list from PROXY_LIST URL
- `api` - Fetch proxy list from API using PROXY_API key

## Environment Variables

Create a `.env` file in the project root with the following variables:

```env
# Proxy list URL for the 'list' command
PROXY_LIST=https://example.com/proxy-list.txt

# API key for the 'api' command (Webshare.io API)
PROXY_API=your_api_key_here
```

## Examples

```bash
# List proxies from URL
./go-proxy list

# Fetch proxies from API
./go-proxy api
```

## Building

```bash
go build -o go-proxy main.go
``` 