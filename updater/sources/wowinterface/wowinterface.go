package wowinterface

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/unly/wow-addon-updater/updater"
	"github.com/unly/wow-addon-updater/updater/sources"
	"github.com/unly/wow-addon-updater/util"
)

var (
	regex = regexp.MustCompile(`^(https?://)?(www\.)?wowinterface\.com/downloads/info.+\.html$`)
)

// source is the source for addons and UIs hosted on wowinterface.com
type source struct {
	downloader sources.Downloader
	client     *http.Client
	baseURL    string
}

//New returns a new update source for wowinterface.com
func New(client *http.Client) (updater.UpdateSource, error) {
	if client == nil {
		client = http.DefaultClient
	}

	d, err := sources.NewDownloader(client)
	if err != nil {
		return nil, err
	}

	return &source{
		downloader: d,
		client:     client,
		baseURL:    "https://www.wowinterface.com",
	}, nil
}

func (source) GetURLRegex() *regexp.Regexp {
	return regex
}

// GetLatestVersion returns the latest version for the given addon URL
func (s *source) GetLatestVersion(addonURL string) (string, error) {
	doc, err := util.GetHTMLPage(s.client, addonURL)
	if err != nil {
		return "", err
	}

	text := doc.Find("#version").Text()
	if !strings.HasPrefix(text, "Version: ") {
		return "", fmt.Errorf("failed to find a version tag for: %s", addonURL)

	}

	return text[9:], nil
}

// DownloadAddon downloads and unzip the addon from the given URL to the given directory
func (s *source) DownloadAddon(addonURL, dir string) error {
	elems := strings.Split(addonURL, "/")
	if len(elems) == 0 {
		return fmt.Errorf("no path to extract from: %s", addonURL)
	}

	name := elems[len(elems)-1]
	name = name[4 : len(name)-5]

	doc, err := util.GetHTMLPage(s.client, fmt.Sprintf("%s/downloads/download%s", s.baseURL, name))
	if err != nil {
		return err
	}

	link, available := doc.Find(".manuallink > a").Attr("href")
	if !available {
		return fmt.Errorf("failed to find download link for: %s", addonURL)
	}

	zipPath, err := s.downloader.DownloadZip(link)
	if err != nil {
		return err
	}

	_, err = util.Unzip(zipPath, dir)
	if err != nil {
		return err
	}

	return nil
}

func (s *source) Close() error {
	return s.downloader.Close()
}
