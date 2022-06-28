package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TNTParser struct {
	errorChan chan error
}

func (parser TNTParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	// Collect non-fatal errors into a slice to be fed into event listener
	var errs []error
	attributes := map[string]string{}
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		fmt.Printf("Whoopsie in TNT parse: %v\n", err)
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	attributes["title"] = getAttributeFromHtmlBasic(doc, "h1", &errs)
	attributes["price"] = strings.ReplaceAll(getAttributeFromHtmlBasic(doc, "d-flex flex-column", &errs), "$", "")
	if attributes["price"] == "" {
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
	}
	//todo: pic
	return attributes, errs
}
