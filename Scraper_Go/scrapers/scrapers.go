package scrapers

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

// Custom user agent.
const (
	userAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) " +
		"AppleWebKit/537.36 (KHTML, like Gecko) " +
		"Chrome/53.0.2785.143 " +
		"Safari/537.36"
)

type WebScraper struct{
	Name string
	sheetsSvc *sheets.Service
	spreadsheetID string
}

type ScraperConfig struct{
	Name				string
	Scope               []string
	CredentialsFilePath string
	SpreadsheetID       string
}

func NewScraper(scraperConfig ScraperConfig) (WebScraper, error){
	ctx := context.Background()
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
	if err != nil{
		return WebScraper{}, fmt.Errorf("unable to retrieve Sheets client: %v", err)
	}

	srv, err := sheets.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return WebScraper{}, fmt.Errorf("unable to retrieve Sheets service: %v", err)
	}

	return WebScraper{
		Name: scraperConfig.Name,
		sheetsSvc: srv,
		spreadsheetID: scraperConfig.SpreadsheetID,
	}, nil
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *jwt.Config) (*http.Client, error) {
	return config.Client(context.Background()), nil
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (s *WebScraper) ScrapeProducts() error{
	fmt.Printf("Starting Scraper: %s\n", s.Name)
	return nil
}
