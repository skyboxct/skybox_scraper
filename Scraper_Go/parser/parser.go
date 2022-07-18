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
	ProductFieldColumns map[string]int
}

func NewProductParser(host string) (iParser, error) {
	//Urls will sometimes be input without 'www.', remove to keep output consistent and absolve reasonable human error
	switch strings.ReplaceAll(host, "www.", "") {
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
