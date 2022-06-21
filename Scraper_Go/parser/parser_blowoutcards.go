package parser

import (
	"fmt"
	"io"
)

type BCParser struct {
}

func (parser BCParser) ParseProductPage(page io.ReadCloser) (map[string]string, error) {
	return nil, fmt.Errorf("Not Implemented!")
}
