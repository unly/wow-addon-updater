package sources

import (
	"io"
	"net/http"
	"os"

	"github.com/unly/wow-addon-updater/util"
)

type Downloader interface {
	io.Closer

	DownloadZip(url string) (string, error)
}

type source struct {
	client  *http.Client
	tempDir string
}

func NewDownloader(client *http.Client) (Downloader, error) {
	path, err := os.MkdirTemp("", "wow-updater")
	if err != nil {
		return nil, err
	}

	if client == nil {
		client = http.DefaultClient
	}

	return &source{
		tempDir: path,
		client:  client,
	}, nil
}

func (s *source) DownloadZip(url string) (string, error) {
	resp, err := s.client.Get(url)
	err = util.CheckHTTPResponse(resp, err)
	if err != nil {
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

func (s *source) Close() error {
	return os.RemoveAll(s.tempDir)
}
