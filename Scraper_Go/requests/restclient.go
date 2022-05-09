package requests

import (
	"net/http"
	"time"
)

type RestClient struct {
	HTTPClient interface {
		Do(*http.Request) (*http.Response, error)
	}
}

func (c *RestClient) Do(req *http.Request) (*http.Response, error) {
	httpClient := c.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	}
	return httpClient.Do(req)
}

type MockClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}


