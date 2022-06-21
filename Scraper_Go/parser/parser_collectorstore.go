package parser

import (
	"fmt"
	"io"
)

type CSParser struct {
}

func (parser CSParser) ParseProductPage(page io.ReadCloser) (map[string]string, error) {
	return nil, fmt.Errorf("Not Implemented!")
}
