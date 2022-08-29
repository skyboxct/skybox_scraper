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
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	//Title
	attributes["title"] = getAttributeFromHtmlBasic(doc, "h1", &errs)

	//Price: check for sale item and add non-sale price
	price := strings.ReplaceAll(doc.Find("span.large").Text(), "$", "")
	if len(price) == 0 {
		//product not on sale, use normal price field
		price = strings.ReplaceAll(doc.Find("strong.large").Text(), "$", "")
	}
	if len(price) == 0 {
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
	}
	attributes["price"] = price

	//Description
	attributes["description"] = getAttributeFromHtmlBasic(doc, "#moredetailsTab > div:nth-child(2) > div:nth-child(1)", &errs)

	//Pic
	var exists bool
	attributes["pic"], exists = doc.Find(".product-image > div:nth-child(1) > a:nth-child(1) > img:nth-child(1)").Attr("src")
	if !exists {
		errs = append(errs, fmt.Errorf("pic not found in html"))
	}

	//UPC
	attributes["upc"] = strings.ReplaceAll(getAttributeFromHtmlBasic(doc, "ul.disc:nth-child(1) > li:nth-child(4)", &errs), "UPC/Barcode: 8", "")

	return attributes, nil
}
