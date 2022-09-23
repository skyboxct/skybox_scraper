package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TWParser struct {
	errorChan chan error
}

func (parser TWParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	// Collect non-fatal errors into a slice to be fed into event listener
	var errs []error
	attributes := map[string]string{}
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	attributes["title"] = getAttributeFromHtmlBasic(doc, "div.productTitle:nth-child(1) > h1:nth-child(1)", &errs)
	//Price: check for sale item and add non-sale price
	price := strings.ReplaceAll(doc.Find(".pvPrice > span:nth-child(1)").Text(), "$", "")
	if len(price) == 0 {
		//product not on sale, use normal price field
		price = strings.ReplaceAll(doc.Find(".pvPrice").Text(), "$", "")
	}
	attributes["price"] = stripPrice(price)
	attributes["stock text"] = strings.ReplaceAll(getAttributeFromHtmlBasic(doc, ".pvDetails > div:nth-child(1) > span:nth-child(3)", &errs), "!", "")
	if attributes["price"] == "" {
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
	}

	var exists bool
	attributes["pic"], exists = doc.Find("#productThumb").Attr("src")
	if !exists {
		errs = append(errs, fmt.Errorf("pic not found in html"))
	}

	return attributes, errs
}
