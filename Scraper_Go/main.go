package main

import "fmt"

// Custom user agent.
const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
	workers = 5
	max_retries = 5
)

func main() {
	fmt.Println("Starting Scraper")
}


