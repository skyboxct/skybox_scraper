package parser

import (
	"fmt"
	"io"
)

type TNTParser struct {
}

func (parser TNTParser) ParseProductPage(page io.ReadCloser) (map[string]string, error) {
	return nil, fmt.Errorf("Not Implemented!")
}
