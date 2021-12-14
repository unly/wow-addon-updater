package util

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/PuerkitoBio/goquery"
)

// CheckHTTPResponse checks the http.Response pointer and error returned
// by the http.RoundTripper. Returns nil if there is no error and the
// status code of the response is 2xx.
func CheckHTTPResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}

	if resp == nil {
		return errors.New("response is nil")
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return fmt.Errorf("http request failed. error code: %d", resp.StatusCode)
	}

	return nil
}

// GetHTMLPage gets the HTML page of the given url and parses the document
// or returns and an error if the HTTP call or the parsing failed.
func GetHTMLPage(client *http.Client, url string) (*goquery.Document, error) {
	resp, err := client.Get(url)
	err = CheckHTTPResponse(resp, err)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}
