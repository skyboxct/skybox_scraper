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
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6)"
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
	Name                string         `json:"name"`
	Scope               []string       `json:"scope"`
	CredentialsFilePath string         `json:"credentials_file_path"`
	SpreadsheetID       string         `json:"spreadsheet_id"`
	ProductSheetName    string         `json:"product_sheet_name"`
	ProductAttributeMap map[string]int `json:"product_attribute_map"`
	Enabled             bool           `json:"enabled"`

	ScraperEventChan chan ScraperEvent `json:"-"`
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

func (s *WebScraper) ScrapeProducts(rowsToInclude []int) error {
	// Todo: split into functions
	if len(rowsToInclude) > 0 {
		fmt.Println("Row override enabled!")
	}

	//Fetch spreadsheet data
	productSpreadsheet, err := s.sheetsSvc.FetchSpreadsheet(s.spreadsheetID)
	if err != nil {
		s.scraperEventChan <- ScraperEvent{
			Level:   FatalError,
			Message: fmt.Sprintf("failed to fetch spreadsheet '%s': %v", s.spreadsheetID, err),
			Scraper: s.Name,
		}
	}

	productSheet, err := productSpreadsheet.SheetByTitle(s.productSheetName)
	if err != nil {
		s.scraperEventChan <- ScraperEvent{
			Level:   FatalError,
			Message: fmt.Sprintf("failed to get sheet '%s': %v", s.productSheetName, err),
			Scraper: s.Name,
		}
	}

	//Fetch all cells containing a product URL
	var urlCells []spreadsheet.Cell
	for _, column := range productSheet.Columns {
		if strings.Contains(column[0].Value, "url") {
			for _, cell := range column[1:] {
				// Cell has a url and is not excluded by provided arguments
				if len(cell.Value) > 0 && (len(rowsToInclude) == 0 || sliceContains(rowsToInclude, int(cell.Row))) {
					urlCells = append(urlCells, cell)
				}
			}
		}
	}

	var wg sync.WaitGroup
	wg.Add(len(urlCells))
	for _, urlCell := range urlCells {
		func(cell spreadsheet.Cell, s *WebScraper) {
			fmt.Printf("CELL: %v\n", cell)
			defer wg.Done()

			url, err := netUrl.Parse(cell.Value)
			if err != nil {
				s.scraperEventChan <- ScraperEvent{
					Level:   ScraperError,
					Message: err.Error(),
					Scraper: s.Name,
					Cell:    cell,
				}
				return
			}

			response, err := s.fetchUrl(url.String())
			if err != nil {
				s.scraperEventChan <- ScraperEvent{
					Level:   ScraperError,
					Message: err.Error(),
					Scraper: s.Name,
					Cell:    cell,
				}
				return
			}

			productParser, err := parser.NewProductParser(url.Host)
			if err != nil {
				s.scraperEventChan <- ScraperEvent{
					Level:   ScraperError,
					Message: err.Error(),
					Scraper: s.Name,
					Cell:    cell,
				}
				return
			}

			productDetails, errs := productParser.ParseProductPage(response)
			for _, err := range errs {
				s.scraperEventChan <- ScraperEvent{
					Level:   ScraperError,
					Message: err.Error(),
					Scraper: s.Name,
					Cell:    cell,
				}
			}

			for attribute, value := range productDetails {
				// Todo: update sheet and push
				//productSheet.Update(int(cell.Row), getAttributeColumn(cell.Value, attribute), value)
				fmt.Printf("%s | %s\n", attribute, value)
			}
			fmt.Println()
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

	if err != nil || resp.StatusCode != 200 || resp.Body == nil {
		return nil, fmt.Errorf("GET failed for %s: %s %v", url, resp.Status, err)
	}

	return resp.Body, nil
}

func sliceContains(s []int, n int) bool {
	for _, v := range s {
		if v == n {
			return true
		}
	}

	return false
}
