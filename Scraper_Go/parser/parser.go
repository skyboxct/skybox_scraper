package parser

import (
	"fmt"
	"io"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

//Todo: Make mapping of product attributes and their corresponding cells
var productAttributeCellsSports map[string]string = map[string]string{
	"dacardworld price": "B",
}

type iParser interface {
	ParseProductPage(page io.ReadCloser) (map[string]string, []error)
	// todo: Get attribute positions?
	// GetAttributeColumn(attribute string)(int)
}

type ProductParser struct {
	ProductFieldColumns map[string]string
}

// Todo: sale prices for all parsers
func NewProductParser(host string) (iParser, error) {
	//Urls will sometimes be input without 'www.', remove to keep output consistent and absolve reasonable human error
	switch host {
	case "dacardworld.com":
		return DAParser{}, nil
	case "steelcitycollectibles.com":
		return SCParser{}, nil
	case "blowoutcards.com":
		return BCParser{}, nil
	case "tcgplayer.com":
		return TCGParser{}, nil
	case "trollandtoad.com":
		return TNTParser{}, nil
	case "toywiz.com":
		return TWParser{}, nil
	case "collectorstore.com":
		return CSParser{}, nil
	default:
		return nil, fmt.Errorf("unrecognized host: %v", host)
	}
}

func getAttributeFromHtmlBasic(doc *goquery.Document, selector string, errorSlice *[]error) string {
	result := doc.Find(selector).Text()
	if len(result) == 0 {
		*errorSlice = append(*errorSlice, fmt.Errorf("%s not found in html", selector))
		return ""
	}
	return result
}

// Removes '$' and ',' from prices to format as a number in sheets
func stripPrice(priceString string) string {
	return strings.ReplaceAll(strings.ReplaceAll(priceString, "$", ""), ",", "")
}
