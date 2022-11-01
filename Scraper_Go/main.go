package main

import (
	"encoding/json"
	"fmt"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"scraper/scrapers"
)

const (
	configFilePath = "scraper_config.json"
)

// TODO: Terminal (simple gui) input for scraper/row config
// TODO: Run nightly on server
// TODO: Error report file

func main() {
	startTime := time.Now()
	eventChan := make(chan scrapers.ScraperEvent)

	go listenForEvents(eventChan)

	var scraperConfigs []*scrapers.ScraperConfig
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

	guiDisabled := false
	if len(os.Args) > 1 {
		for _, arg := range os.Args {
			if arg == "-nogui" {
				guiDisabled = true
			}
		}
	}

	if guiDisabled {
		if len(os.Args) > 1 {
			for _, scraperConfig := range scraperConfigs {
				scraperConfig.RowsToInclude = strings.Join(os.Args, " ")
			}
		}
		registeredScrapers := registerScrapersFromConfigs(scraperConfigs, &eventChan)
		startScraper(registeredScrapers, startTime)
	} else {
		buildAndRunGui(scraperConfigs, startTime, &eventChan)
	}
}

func registerScrapersFromConfigs(scraperConfigs []*scrapers.ScraperConfig, eventChan *chan scrapers.ScraperEvent) []scrapers.WebScraper {
	var registeredScrapers []scrapers.WebScraper
	for _, scraperConfig := range scraperConfigs {
		if scraperConfig.Enabled {
			scraperConfig.ScraperEventChan = *eventChan
			scraper, err := scrapers.NewScraper(*scraperConfig)
			if err != nil {
				fmt.Printf("Error: Failed to initialize %s scraper: %v\n", scraperConfig.Name, err)
			} else {
				registeredScrapers = append(registeredScrapers, scraper)
			}
		}
	}

	return registeredScrapers
}

// TODO: Display all scrapers
// TODO: Add Output
func buildAndRunGui(scraperConfigs []*scrapers.ScraperConfig, startTime time.Time, eventChan *chan scrapers.ScraperEvent) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Skybox Scraper")
	vBoxes := []*fyne.Container{}

	for _, scraperConfig := range scraperConfigs {
		label := widget.NewLabel(scraperConfig.Name)
		enabledBox := widget.NewCheckWithData("Enabled", binding.BindBool(&scraperConfig.Enabled))
		rowOverrideInput := widget.NewEntryWithData(binding.BindString(&scraperConfig.RowsToInclude))
		rowOverrideInput.SetPlaceHolder("Rows to include (leave empty for all) Ex \"1 2-5 20\"")
		vBoxes = append(vBoxes, container.NewVBox(label, enabledBox, rowOverrideInput))
	}

	startButton := widget.NewButton("Start Scraper", func() {
		fmt.Println(scraperConfigs[0].Enabled)
		fmt.Println(&scraperConfigs)
		registeredScrapers := registerScrapersFromConfigs(scraperConfigs, eventChan)
		startScraper(registeredScrapers, startTime)
	})

	myWindow.SetContent(container.NewVBox(vBoxes[0], startButton))
	myWindow.ShowAndRun()
}

func startScraper(registeredScrapers []scrapers.WebScraper, startTime time.Time) {
	var wg sync.WaitGroup
	for _, scraper := range registeredScrapers {
		wg.Add(1)
		go func(webScraper scrapers.WebScraper) {
			fmt.Printf("Starting %s Scraper\n", webScraper.Name)
			err := webScraper.ScrapeProducts()
			if err != nil {
				fmt.Printf("Error in %s Scraper: %v\n", webScraper.Name, err)
			}
			wg.Done()
		}(scraper)
	}
	wg.Wait()

	endTime := time.Since(startTime)
	fmt.Printf("Scrapers Have Finished!\nExecution Time: %v", endTime)
}

func listenForEvents(eventChan chan scrapers.ScraperEvent) {
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
