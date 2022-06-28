package parser

import (
	"fmt"
	"io"

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
	ProductFieldColumns map[string]int
}

func NewProductParser(host string) (iParser, error) {
	switch host {
	case "www.dacardworld.com":
		return DAParser{}, nil
	case "www.steelcitycollectibles.com":
		return SCParser{}, nil
	case "www.blowoutcards.com":
		return BCParser{}, nil
	case "www.tcgplayer.com":
		return TCGParser{}, nil
	case "www.trollandtoad.com":
		return TNTParser{}, nil
	case "www.toywiz.com":
		return TWParser{}, nil
	default:
		return nil, fmt.Errorf("unrecognized host: %v", host)
	}
}

func getAttributeFromHtmlBasic(doc *goquery.Document, selector string, errorSlice *[]error) string {
	result := doc.Find(selector).Text()
	fmt.Println(result)
	if len(result) == 0 {
		*errorSlice = append(*errorSlice, fmt.Errorf("%s not found in html", selector))
		return ""
	}
	return result
}
