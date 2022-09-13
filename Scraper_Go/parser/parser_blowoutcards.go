package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type BCParser struct {
	errorChan chan error
}

func (parser BCParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	// Collect non-fatal errors into a slice to be fed into event listener
	var errs []error
	attributes := map[string]string{}
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	attributes["title"] = getAttributeFromHtmlBasic(doc, "div.product-name > h1:nth-child(1)", &errs)
	attributes["price"] = strings.ReplaceAll(getAttributeFromHtmlBasic(doc, "div.price-box:nth-child(2)", &errs), "$", "")
	attributes["stock text"] = strings.ReplaceAll(getAttributeFromHtmlBasic(doc, ".availability", &errs), "Availability: ", "")
	var exists bool
	picPath, exists := doc.Find("#zoom1").Attr("href")
	if !exists {
		errs = append(errs, fmt.Errorf("pic not found in html"))
	} else {
		attributes["pic"] = "www.blowoutcards.com" + picPath
	}

	return attributes, errs
}
