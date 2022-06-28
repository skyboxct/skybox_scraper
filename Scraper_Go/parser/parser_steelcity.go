package parser

import (
	"fmt"
	"io"
)

type SCParser struct {
	errorChan chan error
}

func (parser SCParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	return nil, []error{fmt.Errorf("Not Implemented!")}
}
