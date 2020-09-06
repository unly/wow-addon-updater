package sources

import (
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	file, err := ioutil.TempFile(s.tempDir, "*.zip")
	if err != nil {
		return "", err
	}

	path := file.Name()
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)

	return path, err
}

func createTempDir(postfix string) string {
	path, err := ioutil.TempDir("", "wow-updater-"+postfix)
	if err != nil {
		panic(err)
	}

	return path
}
