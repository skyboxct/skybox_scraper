package parser

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SCParser struct {
	errorChan chan error
}

func (parser SCParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	var errs []error
	attributes := map[string]string{}

	doc, err := goquery.NewDocumentFromReader(page)
	if err != nil {
		return nil, []error{fmt.Errorf("could not create searchable document from html, %v", err)}
	}

	attributes["title"] = getAttributeFromHtmlBasic(doc, ".five > h1:nth-child(1)", &errs)

	outOfStock := strings.Contains(strings.ToLower(doc.Find(".five > div:nth-child(3) > p:nth-child(4)").Text()), "out of stock")
	if outOfStock {
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
		//Todo: steelCity sale price
		attributes["price"] = getAttributeFromHtmlBasic(doc, ".p-price > span:nth-child(1)", &errs)
	}

	picHtml, err := doc.Find(".seven").Html()
	if err != nil {
		errs = append(errs, fmt.Errorf("error getting html for pic: %v", err))
	}
	r := regexp.MustCompile("https://www.steelcitycollectibles.com/storage/img/uploads/products/full/.*jpg")
	attributes["pic"] = r.FindString(picHtml)

	return attributes, errs
}
