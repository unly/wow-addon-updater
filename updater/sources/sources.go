package sources

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

type source struct {
	regex   *regexp.Regexp
	tempDir string
}

func newSource(regex, name string) *source {
	return &source{
		regex:   regexp.MustCompile(regex),
		tempDir: createTempDir(name),
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

	file, err := ioutil.TempFile(s.tempDir, "*.zip")
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)

	return file.Name(), err
}

func createTempDir(postfix string) string {
	path, err := ioutil.TempDir("", "wow-updater-"+postfix)
	if err != nil {
		panic(err)
	}

	return path
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
