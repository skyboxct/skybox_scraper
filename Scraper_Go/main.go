package main

import (
	"fmt"
	"os"

	"scraper/scrapers"
)

const (
	sportsSpreadsheetID = "1_KRrcUKx42H9Ybq6hFguCsnFyDeQlk9-iI3fe1Pi2lU"
	tcgSpreadsheetID    = "1k4fWKBXWr74A-5s8pQ6PW8ZhcoGe9nogQjjn1syuYyo"

	sportsProductSheetName = "Sports Website Scraper"
	tcgProductSheetName    = "TCG Website Scraper"

	sportsCredsFilePath = "../sportscreds.json"
	tcgCredsFilePath    = "../tcgcreds.json"
)

const (
	workers     = 5
	max_retries = 5
)

func main() {
	var registeredScrapers []scrapers.WebScraper

	// Register Sports Scraper
	if len(os.Args) == 1 || contains(os.Args, "sports") {
		fmt.Println("Init Sports")
		sportsConfig := scrapers.ScraperConfig{
			Name:                "Sports",
			Scope:               []string{"https://spreadsheets.google.com/feeds", "https://www.googleapis.com/auth/drive"},
			SpreadsheetID:       sportsSpreadsheetID,
			CredentialsFilePath: sportsCredsFilePath,
			ProductSheetName:    sportsProductSheetName,
		}
		sportsScraper, err := scrapers.NewScraper(sportsConfig)
		if err != nil {
			fmt.Printf("Error: Failed to initialize sports scraper: %v\n", err)
		} else {
			registeredScrapers = append(registeredScrapers, sportsScraper)
		}
	}

	// Register TCG Scraper
	if len(os.Args) == 1 || contains(os.Args, "tcg") {
		fmt.Println("Init TCG")
		tcgConfig := scrapers.ScraperConfig{
			Name:                "TCG",
			Scope:               []string{"https://spreadsheets.google.com/feeds", "https://www.googleapis.com/auth/drive"},
			SpreadsheetID:       tcgSpreadsheetID,
			CredentialsFilePath: tcgCredsFilePath,
			ProductSheetName:    tcgProductSheetName,
		}
		tcgScraper, err := scrapers.NewScraper(tcgConfig)
		if err != nil {
			fmt.Printf("Error: Failed to initialize tcg scraper: %v\n", err)
		} else {
			registeredScrapers = append(registeredScrapers, tcgScraper)
		}
	}

	//TODO: Run scrapers concurrently?
	for _, scraper := range registeredScrapers {
		fmt.Printf("Starting %s scraper", scraper.Name)
		err := scraper.ScrapeProducts()
		if err != nil {
			fmt.Printf("Error in %s Scraper: %v\n", scraper.Name, err)
		}
	}
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}
