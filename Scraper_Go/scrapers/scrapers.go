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

type WebScraper struct {
	Name             string
	sheetsSvc        *spreadsheet.Service
	spreadsheetID    string
	productSheetName string
	numWorkers       int
	httpClient       http.Client
}

type ScraperConfig struct {
	Name                string
	Scope               []string
	CredentialsFilePath string
	SpreadsheetID       string
	ProductSheetName    string
}

type Product struct {
	row    uint
	errors []error
	urls   []spreadsheet.Cell
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
	}, nil
}

func getSheetsClient(config *jwt.Config) (*http.Client, error) {
	return config.Client(context.Background()), nil
}

func (s *WebScraper) ScrapeProducts() error {
	//Fetch spreadsheet data
	productSpreadsheet, err := s.sheetsSvc.FetchSpreadsheet(s.spreadsheetID)
	if err != nil {
		return fmt.Errorf("failed to fetch spreadsheet '%s': %v", s.spreadsheetID, err)
	}

	productSheet, err := productSpreadsheet.SheetByTitle(s.productSheetName)
	if err != nil {
		return fmt.Errorf("failed to get sheet '%s': %v", s.productSheetName, err)
	}

	//Fetch all cells containing a product URL
	var urlCells []spreadsheet.Cell
	for _, column := range productSheet.Columns {
		if strings.Contains(column[0].Value, "url") {
			for _, cell := range column {
				if len(cell.Value) > 0 {
					urlCells = append(urlCells, cell)
					fmt.Printf("URL CELL: ROW = %v  COL = %v  VAL = %v\n", cell.Row, cell.Column, cell.Value)
				}
			}
		}
	}

	urlCells = []spreadsheet.Cell{productSheet.Columns[17][1]}
	fmt.Println(urlCells)

	var wg sync.WaitGroup
	chFailures := make(chan error)
	wg.Add(len(urlCells))
	for _, urlCell := range urlCells {
		func(cell spreadsheet.Cell, failChan chan error) {
			defer wg.Done()
			response, err := s.fetchUrl(cell.Value)
			if err != nil {
				fmt.Println(err)
				failChan <- err
			} else {
				url, err := netUrl.Parse(cell.Value)
				if err != nil {
					failChan <- err
				}

				productParser, err := parser.NewProductParser(url.Host)
				if err != nil {
					fmt.Println(err)
					failChan <- err
				}

				productDetails, err := productParser.ParseProductPage(response)
				if err != nil {
					fmt.Println(err)
					failChan <- err
				} else {
					for attribute, value := range productDetails {
						//productSheet.Update(int(cell.Row), getAttributeColumn(cell.Value, attribute), value)
						fmt.Println("ATTRIBUTE AND VALUE: ", attribute, value)
					}
				}
			}
		}(urlCell, chFailures)
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
