package main

import (
	"fmt"
	"os"
	"scraper/scrapers"
)

const (
	sportsSpreadsheetID = ""
	tcgSpreadsheetID = ""
	
	sportsCredsFilePath = "sportscreds.json"
	tcgCredsFilePath = "tcgcreds.json"
)

const(
	workers = 5
	max_retries = 5
)

func main() {
	fmt.Println("Starting Scraper")
	var registeredScrapers []scrapers.WebScraper

	// Register Sports Scraper
	if len(os.Args) == 0 || contains(os.Args, "sports"){
		sportsConfig := scrapers.ScraperConfig{
			Scope: []string{"https://spreadsheets.google.com/feeds", "https://www.googleapis.com/auth/drive"},
			SpreadsheetID: sportsSpreadsheetID,
			CredentialsFilePath: sportsCredsFilePath,
		}
		sportsScraper, err := scrapers.NewScraper(sportsConfig)
		if err != nil{
			fmt.Printf("Error: Failed to initialize sports scraper: %v", err)
		} else{
			registeredScrapers = append(registeredScrapers, sportsScraper)
		}
	}

	// Register TCG Scraper
	if len(os.Args) == 0 || contains(os.Args, "tcg"){
		tcgConfig := scrapers.ScraperConfig{
			Scope: []string{"https://spreadsheets.google.com/feeds", "https://www.googleapis.com/auth/drive"},
			SpreadsheetID: tcgSpreadsheetID,
			CredentialsFilePath: tcgCredsFilePath,
		}
		tcgScraper, err := scrapers.NewScraper(tcgConfig)
		if err != nil{
			fmt.Printf("Error: Failed to initialize tcg scraper: %v", err)
		} else{
			registeredScrapers = append(registeredScrapers, tcgScraper)
		}
	}

	//TODO: Run scrapers concurrently?
	for _, scraper := range registeredScrapers{
		err := scraper.ScrapeProducts()
		if err != nil{
			fmt.Printf("Error in %s Scraper: %v", scraper.Name, err)
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


