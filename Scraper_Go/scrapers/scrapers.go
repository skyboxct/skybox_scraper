package scrapers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"gopkg.in/Iwark/spreadsheet.v2"
)

// Custom user agent.
const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
)

const maxWorkers = 10

type WebScraper struct {
	Name             string
	sheetsSvc        *spreadsheet.Service
	spreadsheetID    string
	productSheetName string
	numWorkers       int
}

type ScraperConfig struct {
	Name                string
	Scope               []string
	CredentialsFilePath string
	SpreadsheetID       string
	ProductSheetName    string
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
	client, err := getClient(config)
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
	}, nil
}

func getClient(config *jwt.Config) (*http.Client, error) {
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
			urlCells = append(urlCells, column[1:]...)
		}
	}

	//Feed cells into parsers asynchronously
	//TODO: Not suck
	var wg sync.WaitGroup
	for i := 0; i < len(urlCells); i += s.numWorkers {
		wg.Add(s.numWorkers)
		for j := i; j < i+s.numWorkers; j++ {
			go func() {
				defer wg.Done()
				fmt.Println(urlCells[j].Value)
				time.Sleep(1)
			}()
		}
		wg.Wait()
	}

	return nil
}
