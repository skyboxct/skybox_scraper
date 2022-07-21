package parser

import (
	"fmt"
	"io"
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
		attributes["price"] = getAttributeFromHtmlBasic(doc, ".p-price > span:nth-child(1)", &errs)
	}

	var exists bool
	//str, _ := doc.Html()
	//fmt.Println(str)
	attributes["pic"], exists = doc.Find(".lum-img").Attr("src")
	if !exists {
		errs = append(errs, fmt.Errorf("pic not found in html"))
	}

	return attributes, errs
}
