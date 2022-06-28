package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TCGParser struct {
	errorChan chan error
}

func (parser TCGParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	// Collect non-fatal errors into a slice to be fed into event listener
	var errs []error
	attributes := map[string]string{}
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	attributes["title"] = getAttributeFromHtmlBasic(doc, ".product-details__name", &errs)
	attributes["price"] = strings.ReplaceAll(getAttributeFromHtmlBasic(doc, ".spotlight__price", &errs), "$", "")
	if attributes["price"] == "" {
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
	}
	attributes["description"] = getAttributeFromHtmlBasic(doc, ".pd-description__description", &errs)
	//todo: pic
	return attributes, errs
}
