package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"

	"scraper/scrapers"
)

const (
	configFilePath = "scraper_config.json"
)

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
			if strings.ToLower(arg) == "-nogui" {
				guiDisabled = true
			}
		}
	}

	if guiDisabled {
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

// TODO: Add Output
func buildAndRunGui(scraperConfigs []*scrapers.ScraperConfig, startTime time.Time, eventChan *chan scrapers.ScraperEvent) {
	myApp := app.New()
	myWindow := myApp.NewWindow("Skybox Scraper")
	myWindow.Resize(fyne.Size{
		Width:  600,
		Height: 100,
	})
	vBoxes := []*fyne.Container{}

	for _, scraperConfig := range scraperConfigs {
		label := widget.NewLabel(scraperConfig.Name)
		enabledBox := widget.NewCheckWithData("Enabled", binding.BindBool(&scraperConfig.Enabled))
		rowOverrideInput := widget.NewEntryWithData(binding.BindString(&scraperConfig.RowsToInclude))
		rowOverrideInput.MultiLine = true
		rowOverrideInput.SetPlaceHolder("Rows to include, space-separated\n(leave empty for all)\nEx \"1 2-5 20\"")
		//vBoxes = append(vBoxes, *container.New(layout.NewVBoxLayout(), label, enabledBox, rowOverrideInput))
		vBoxes = append(vBoxes, container.NewVBox(label, enabledBox, rowOverrideInput))
	}

	configBoxes := container.New(layout.NewGridLayout(len(vBoxes)), vBoxes[0], vBoxes[1])

	startButton := widget.NewButton("Start Scraper", func() {
		registeredScrapers := registerScrapersFromConfigs(scraperConfigs, eventChan)
		startScraper(registeredScrapers, startTime)
	})

	//widget.NewLabelWithData(binding.BindString())
	outputBox := container.New(layout.NewPaddedLayout())

	myWindow.SetContent(container.New(layout.NewVBoxLayout(), configBoxes, startButton, outputBox))
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
