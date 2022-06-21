package parser

import (
	"fmt"
	"io"
)

//Todo: Make mapping of product attributes and their corresponding cells
var productAttributeCellsSports map[string]string = map[string]string{
	"dacardworld price": "B",
}

type iParser interface {
	ParseProductPage(page io.ReadCloser) (map[string]string, error)
	// todo: Get attribute positions?
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
