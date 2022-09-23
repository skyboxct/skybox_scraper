package parser

import (
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
)

type TCGParser struct {
	errorChan chan error
}

func (parser TCGParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	//Todo: Rewrite to match new TCG page format
	return nil, []error{fmt.Errorf("TCG PLAYER CURRENTLY UNAVAILABLE")}

	// Collect non-fatal errors into a slice to be fed into event listener
	var errs []error
	attributes := map[string]string{}
	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	attributes["title"] = getAttributeFromHtmlBasic(doc, ".product-details__name", &errs)
	attributes["description"] = getAttributeFromHtmlBasic(doc, ".product__item-details__description", &errs)
	attributes["price"] = stripPrice(getAttributeFromHtmlBasic(doc, ".spotlight__price", &errs))
	if attributes["price"] == "" {
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
	}

	var exists bool
	attributes["pic"], exists = doc.Find(".progressive-image-main").Attr("src")
	if !exists {
		errs = append(errs, fmt.Errorf("pic not found in html"))
	}
	return attributes, errs
}
