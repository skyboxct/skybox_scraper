package parser

import (
	"fmt"
	"io"

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
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	attributes["title"] = getAttributeFromHtmlBasic(doc, "h1.font-weight-bold", &errs)
	attributes["price"] = stripPrice(getAttributeFromHtmlBasic(doc, ".d-lg-block > div:nth-child(1) > div:nth-child(1) > div:nth-child(1) > div:nth-child(1) > div:nth-child(1) > span:nth-child(1)", &errs))
	if attributes["price"] == "" {
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
	}

	var exists bool
	attributes["pic"], exists = doc.Find("img.mw-100").Attr("src")
	if !exists {
		errs = append(errs, fmt.Errorf("pic not found in html"))
	}

	return attributes, errs
}
