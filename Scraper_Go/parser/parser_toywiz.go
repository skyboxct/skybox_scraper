package parser

import (
	"fmt"
	"io"
)

type TWParser struct {
}

func (parser TWParser) ParseProductPage(page io.ReadCloser) (map[string]string, error) {
	return nil, fmt.Errorf("Not Implemented!")
}
