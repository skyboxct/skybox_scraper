package parser

import (
	"fmt"
	"io"
)

type BCParser struct {
	errorChan chan error
}

func (parser BCParser) ParseProductPage(page io.ReadCloser) (map[string]string, []error) {
	return nil, []error{fmt.Errorf("Not Implemented!")}
}
