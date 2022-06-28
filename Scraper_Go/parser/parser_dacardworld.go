package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type DAParser struct {
	errorChan chan error
}

func (parser DAParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	// Collect non-fatal errors into a slice to be fed into event listener
	var errs []error
	attributes := map[string]string{}
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		fmt.Printf("Whoopsie in DA parse: %v\n", err)
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	attributes["title"] = getAttributeFromHtmlBasic(doc, "h1", &errs)

	//Get Price: check for sale item and add non-sale price
	price := strings.ReplaceAll(doc.Find("price discount large").Text(), "$", "")
	if len(price) == 0 {
		//product not on sale, use normal price field
		price = doc.Find("price large").Text()
	}
	if len(price) == 0 {
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
	}
	attributes["price"] = price

	//Get Price: check for sale item and add non-sale price
	description := doc.Find("eight columns").Text()
	if len(description) == 0 {
		//product not on sale, use normal price field
		price = doc.Find("moredetailsTab").Text()
	}
	if len(description) == 0 {
		errs = append(errs, fmt.Errorf("description not found in html"))
	}
	attributes["description"] = description

	//todo: pic
	//todo: upc
	return attributes, nil
}
