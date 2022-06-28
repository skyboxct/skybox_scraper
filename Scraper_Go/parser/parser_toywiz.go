package parser

import (
	"fmt"
	"io"
)

type TWParser struct {
	errorChan chan error
}

func (parser TWParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	return nil, []error{fmt.Errorf("Not Implemented!")}
}
