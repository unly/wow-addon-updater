package sources

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

type source struct {
	regex   *regexp.Regexp
	tempDir string
}

func newSource(regex *regexp.Regexp) *source {
	path, err := os.MkdirTemp("", "wow-updater")
	if err != nil {
		panic(err)
	}

	return &source{
		regex:   regex,
		tempDir: path,
	}
}

func (s *source) GetURLRegex() *regexp.Regexp {
	return s.regex
}

func (s *source) downloadZip(url string) (string, error) {
	resp, err := http.Get(url)
	if err := checkHTTPResponse(resp, err); err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := os.CreateTemp(s.tempDir, "*.zip")
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)

	return file.Name(), err
}

func (s *source) Close() {
	_ = os.RemoveAll(s.tempDir)
}

func checkHTTPResponse(resp *http.Response, err error) error {
	if err != nil {
		return err
	}

	if resp == nil {
		return nil
	}

	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return fmt.Errorf("http request failed. error code: %s", resp.Status)
	}

	return nil
}
