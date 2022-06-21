package parser

import (
	"fmt"
	"io"

	"github.com/PuerkitoBio/goquery"
)

type DAParser struct {
}

func (parser DAParser) ParseProductPage(page io.ReadCloser) (map[string]string, error) {
	attributes := map[string]string{}
	doc, err := goquery.NewDocumentFromReader(page)

	if err != nil {
		fmt.Printf("Whoopsie in DA parse: %v\n", err)
	}

	//todo: make helper function w/ error channel
	title := doc.Find("h1").Text()
	fmt.Println(title)
	if len(title) == 0 {
		fmt.Println("TITLE NOT FOUND")
	}
	attributes["title"] = title

	//Get Price: check for sale item and add non-sale price
	//todo: remove $
	price := doc.Find("price discount large").Text()
	if len(price) == 0 {
		//product not on sale, use normal price field
		price = doc.Find("price large").Text()
	}
	if len(price) == 0 {
		fmt.Println("PRICE NOT FOUND")
		attributes["stock text"] = "Out of Stock"
	} else {
		attributes["stock text"] = "In Stock"
	}
	attributes["price"] = price

	//Get Price: check for sale item and add non-sale price
	description := doc.Find("eight columns").Text()
	if len(description) == 0 {
		//product not on sale, use normal price field
		price = doc.Find("moredetailsTab").Text()
	}
	if len(description) == 0 {
		fmt.Println("DESCRIPTION NOT FOUND")
	}
	attributes["description"] = description

	//todo: pic
	//todo: upc
	return attributes, nil
}
