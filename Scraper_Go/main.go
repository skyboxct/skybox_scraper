package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"scraper/scrapers"
)

const (
	configFilePath = "scraper_config.json"
)

func main() {
	var registeredScrapers []scrapers.WebScraper
	var rowsToInclude []int
	eventChan := make(chan scrapers.ScraperEvent)

	go listenForEvents(eventChan)

	// TODO: Make hosts configurable
	var scraperConfigs []scrapers.ScraperConfig
	configFileContents, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("[ERROR] Failed to read config file %s: %v\nExiting...", configFilePath, err)
		os.Exit(1)
	}

	err = json.Unmarshal(configFileContents, &scraperConfigs)
	if err != nil {
		fmt.Printf("[ERROR] Failed to marshal config file %s: %v\nExiting...", configFilePath, err)
		os.Exit(1)
	}

	for _, scraperConfig := range scraperConfigs {
		if scraperConfig.Enabled {
			scraperConfig.ScraperEventChan = eventChan
			scraper, err := scrapers.NewScraper(scraperConfig)
			if err != nil {
				fmt.Printf("Error: Failed to initialize %s scraper: %v\n", scraperConfig.Name, err)
			} else {
				registeredScrapers = append(registeredScrapers, scraper)
			}
		}
	}

	if len(os.Args) > 1 {
		for _, arg := range os.Args {
			if row, err := strconv.Atoi(arg); err == nil {
				rowsToInclude = append(rowsToInclude, row-1)
			} else if strings.Contains(arg, "-") {
				bounds := strings.Split(arg, "-")
				if len(bounds) != 2 {
					fmt.Printf("[ERROR] Invalid row range input: %s\n", arg)
					os.Exit(1)
				}

				start, err := strconv.Atoi(bounds[0])
				end, err := strconv.Atoi(bounds[1])
				if err != nil || end < start {
					fmt.Printf("[ERROR] Invalid row range input: %s\n", arg)
					os.Exit(1)
				}

				for i := start - 1; i <= end-1; i++ {
					rowsToInclude = append(rowsToInclude, i)
				}
			}
		}
	}

	//TODO: Run scrapers concurrently?
	for _, scraper := range registeredScrapers {
		fmt.Printf("Starting %s Scraper\n", scraper.Name)
		err := scraper.ScrapeProducts(rowsToInclude)
		if err != nil {
			fmt.Printf("Error in %s Scraper: %v\n", scraper.Name, err)
		}
	}
}

func listenForEvents(eventChan chan scrapers.ScraperEvent) {
	//todo: collect error cells, retry mechanism
	for {
		select {
		case event := <-eventChan:
			switch event.Level {
			case scrapers.Info:
				fmt.Printf("[INFO] Scraper: %s ** Message: %s\n", event.Scraper, event.Message)
			case scrapers.Warning:
				fmt.Printf("[WARN] Scraper: %s ** Message: %s\n", event.Scraper, event.Message)
			case scrapers.ScraperError:
				fmt.Printf("[ERROR] Scraper: %s ** Message: %s\n Cell impacted:%v\n", event.Scraper, event.Message, event.Cell)
			case scrapers.FatalError:
				fmt.Printf("[FATAL] Error in %s Scraper!! %s\n Cell Impacted: %v\n This is a fatal error, scraper will shut down", event.Scraper, event.Message, event.Cell)
				os.Exit(1)
			}
		}
	}
}
