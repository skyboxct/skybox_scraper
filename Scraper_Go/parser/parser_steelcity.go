package parser

import (
	"fmt"
	"io"
)

type SCParser struct {
}

func (parser SCParser) ParseProductPage(page io.ReadCloser) (map[string]string, error) {
	return nil, fmt.Errorf("Not Implemented!")
}
