package parser

import (
	"fmt"
	"io"
)

type TCGParser struct {
}

func (parser TCGParser) ParseProductPage(page io.ReadCloser) (map[string]string, error) {
	return nil, fmt.Errorf("Not Implemented!")
}
