package scrapers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewScraper(t *testing.T){
	t.Run("Successfully creates new scraper from config", func(t *testing.T) {
		config := ScraperConfig{
			scope:               []string{"https://testscope.com"},
			credentialsFilePath: "testdata/testcreds.json",
		}
		scraper, err := NewScraper(config)
		assert.NoError(t, err)
		assert.NotNil(t, scraper)
		assert.NotNil(t, scraper.sheetsSvc)
	})
}
