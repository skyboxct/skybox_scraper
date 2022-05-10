package scrapers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScraper(t *testing.T){
	t.Run("Successfully creates new scraper from config", func(t *testing.T) {
		config := ScraperConfig{
			Scope:               []string{"https://testscope.com"},
			CredentialsFilePath: "testdata/testcreds.json",
			SpreadsheetID:       "testID",
			Name: "testName",
		}
		scraper, err := NewScraper(config)
		assert.NoError(t, err)
		assert.NotNil(t, scraper)
		assert.NotNil(t, scraper.sheetsSvc)
		assert.Equal(t, config.SpreadsheetID, scraper.spreadsheetID)
		assert.Equal(t, config.Name, scraper.Name)
	})
}
