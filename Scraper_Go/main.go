package main

import (
	"fmt"
)

const(
	workers = 5
	max_retries = 5
)

func main() {
	fmt.Println("Starting Scraper")
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}


