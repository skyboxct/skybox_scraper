package scrapers

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	netUrl "net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"gopkg.in/Iwark/spreadsheet.v2"

	"scraper/parser"
)

// Custom user agent.
const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
)

const (
	httpTimeout = 60 * time.Second
	maxWorkers  = 10
)

type eventLevel uint8

const (
	Info eventLevel = iota
	Warning
	ScraperError
	FatalError
)

type WebScraper struct {
	Name             string
	sheetsSvc        *spreadsheet.Service
	spreadsheetID    string
	productSheetName string
	numWorkers       int
	httpClient       http.Client
	scraperEventChan chan ScraperEvent
}

type ScraperConfig struct {
	Name                string
	Scope               []string
	CredentialsFilePath string
	SpreadsheetID       string
	ProductSheetName    string
	ScraperEventChan    chan ScraperEvent
}

type ScraperEvent struct {
	Level   eventLevel
	Message string
	Scraper string
	Cell    spreadsheet.Cell
}

func NewScraper(scraperConfig ScraperConfig) (WebScraper, error) {
	b, err := ioutil.ReadFile(scraperConfig.CredentialsFilePath)
	if err != nil {
		return WebScraper{}, fmt.Errorf("unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved token.json.
	config, err := google.JWTConfigFromJSON(b, scraperConfig.Scope...)
	if err != nil {
		return WebScraper{}, fmt.Errorf("unable to parse client secret file to config: %v", err)
	}
	client, err := getSheetsClient(config)
	if err != nil {
		return WebScraper{}, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	srv := spreadsheet.NewServiceWithClient(client)

	return WebScraper{
		Name:             scraperConfig.Name,
		sheetsSvc:        srv,
		spreadsheetID:    scraperConfig.SpreadsheetID,
		productSheetName: scraperConfig.ProductSheetName,
		numWorkers:       maxWorkers,
		httpClient:       http.Client{Timeout: httpTimeout},
		scraperEventChan: scraperConfig.ScraperEventChan,
	}, nil
}

func getSheetsClient(config *jwt.Config) (*http.Client, error) {
	return config.Client(context.Background()), nil
}

func (s *WebScraper) ScrapeProducts() error {
	//Fetch spreadsheet data
	productSpreadsheet, err := s.sheetsSvc.FetchSpreadsheet(s.spreadsheetID)
	if err != nil {
		s.scraperEventChan <- ScraperEvent{
			Level:   FatalError,
			Message: fmt.Sprintf("failed to fetch spreadsheet '%s': %v", s.spreadsheetID, err),
			Scraper: s.Name,
		}
		//return fmt.Errorf("failed to fetch spreadsheet '%s': %v", s.spreadsheetID, err)
	}

	productSheet, err := productSpreadsheet.SheetByTitle(s.productSheetName)
	if err != nil {
		s.scraperEventChan <- ScraperEvent{
			Level:   FatalError,
			Message: fmt.Sprintf("failed to get sheet '%s': %v", s.productSheetName, err),
			Scraper: s.Name,
		}
		//return fmt.Errorf("failed to get sheet '%s': %v", s.productSheetName, err)
	}

	//Fetch all cells containing a product URL
	var urlCells []spreadsheet.Cell
	for _, column := range productSheet.Columns {
		if strings.Contains(column[0].Value, "url") {
			for _, cell := range column[1:] {
				if len(cell.Value) > 0 {
					urlCells = append(urlCells, cell)
				}
			}
		}
	}

	urlCells = []spreadsheet.Cell{productSheet.Columns[17][1]}
	s.scraperEventChan <- ScraperEvent{
		Level:   Info,
		Message: fmt.Sprintf("%v", urlCells),
		Scraper: s.Name,
	}

	var wg sync.WaitGroup
	wg.Add(len(urlCells))
	for _, urlCell := range urlCells {
		func(cell spreadsheet.Cell, s *WebScraper) {
			defer wg.Done()
			response, err := s.fetchUrl(cell.Value)
			if err != nil {
				s.scraperEventChan <- ScraperEvent{
					Level:   ScraperError,
					Message: err.Error(),
					Scraper: s.Name,
					Cell:    urlCell,
				}
			} else {
				url, err := netUrl.Parse(cell.Value)
				if err != nil {
					s.scraperEventChan <- ScraperEvent{
						Level:   ScraperError,
						Message: err.Error(),
						Scraper: s.Name,
						Cell:    urlCell,
					}
				}

				productParser, err := parser.NewProductParser(url.Host)
				if err != nil {
					s.scraperEventChan <- ScraperEvent{
						Level:   ScraperError,
						Message: err.Error(),
						Scraper: s.Name,
						Cell:    urlCell,
					}
				}

				productDetails, errs := productParser.ParseProductPage(response)
				for _, err := range errs {
					s.scraperEventChan <- ScraperEvent{
						Level:   ScraperError,
						Message: err.Error(),
						Scraper: s.Name,
						Cell:    urlCell,
					}
				}
				for attribute, value := range productDetails {
					//productSheet.Update(int(cell.Row), getAttributeColumn(cell.Value, attribute), value)
					fmt.Println("ATTRIBUTE AND VALUE: ", attribute, value)
				}
			}
		}(urlCell, s)
	}
	wg.Wait()

	return nil
}

func (s *WebScraper) fetchUrl(url string) (io.ReadCloser, error) {
	// Open url.
	// Need to use http.Client in order to set a custom user agent:
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", userAgent)
	resp, err := s.httpClient.Do(req)

	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}

	return resp.Body, nil
}
