package tukui

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/unly/go-tukui"

	"github.com/unly/wow-addon-updater/updater"
	"github.com/unly/wow-addon-updater/updater/sources"
	"github.com/unly/wow-addon-updater/util"
)

var (
	regex   = regexp.MustCompile(`^(https?://)?(www\.)?tukui\.org/((classic-(tbc-)?)?addons\.php\?id=[0-9]+)|(download\.php\?ui=(tukui|elvui))$`)
	idRegex = regexp.MustCompile(`id=[0-9]+`)
	uiRegex = regexp.MustCompile(`ui=.+`)
)

//go:generate go run github.com/vektra/mockery/v2 --case=underscore  --name=tukuiAPI --structname=MockTukUIAPI

type tukuiAPI interface {
	tukui.AddonClient
}

type tukUISource struct {
	downloader sources.Downloader
	client     *http.Client
	classic    tukuiAPI
	retail     tukuiAPI
}

// New returns a pointer to a newly created TukUISource.
func New(client *http.Client) (updater.UpdateSource, error) {
	if client == nil {
		client = http.DefaultClient
	}
	d, err := sources.NewDownloader(client)
	if err != nil {
		return nil, err
	}

	tukClient := tukui.NewClient(client)

	return &tukUISource{
		downloader: d,
		client:     client,
		classic:    tukClient.ClassicAddons,
		retail:     tukClient.RetailAddons,
	}, nil
}

func (tukUISource) GetURLRegex() *regexp.Regexp {
	return regex
}

// GetLatestVersion returns the latest version for the given addon URL
func (t *tukUISource) GetLatestVersion(addonURL string) (string, error) {
	tukuiAddon, err := t.getAddon(addonURL)
	if err != nil {
		return "", err
	}

	version := tukuiAddon.Version
	if version == nil {
		return "", errors.New("the api response did not contain a version")
	}

	return *version, nil
}

func (t *tukUISource) getAddon(url string) (tukui.Addon, error) {
	// regular addon parameter queries
	if idRegex.FindString(url) != "" {
		return t.getRegularAddon(url)
	}

	// tukui and elvui queries
	if uiRegex.FindString(url) != "" {
		return t.getUIAddon(url)
	}

	return tukui.Addon{}, fmt.Errorf("tukui.org url %s is not supported", url)
}

func (t *tukUISource) getUIAddon(url string) (tukui.Addon, error) {
	uiRunes := []rune(uiRegex.FindString(url))
	if len(uiRunes) < 3 {
		return tukui.Addon{}, fmt.Errorf("failed to extract the ui= parameter from %s. no ui found", url)
	}

	switch string(uiRunes[3:]) {
	case "tukui":
		ui, resp, err := t.retail.GetTukUI()
		return ui, util.CheckHTTPResponse(resp, err)
	case "elvui":
		ui, resp, err := t.retail.GetElvUI()
		return ui, util.CheckHTTPResponse(resp, err)
	default:
		return tukui.Addon{}, fmt.Errorf("given tukui.org ui addon link %s is not supported", url)
	}
}

func (t *tukUISource) getRegularAddon(url string) (tukui.Addon, error) {
	addon := tukui.Addon{}

	doc, err := util.GetHTMLPage(t.client, url)
	if err != nil {
		return addon, err
	}

	s := doc.Find("#extras .extras:nth-of-type(1) > b.VIP:nth-of-type(1)")
	if s == nil || len(s.Text()) == 0 {
		return addon, fmt.Errorf("failed to query %s page for a version", url)
	}

	version := s.Text()
	addon.Version = &version
	downloadURL := strings.Replace(url, "id", "download", 1)
	addon.URL = &downloadURL

	return addon, nil
}

// DownloadAddon downloads and unzip the addon from the given URL to the given directory
func (t *tukUISource) DownloadAddon(addonURL, dir string) error {
	tukuiAddon, err := t.getAddon(addonURL)
	if err != nil {
		return err
	}

	url := tukuiAddon.URL
	if url == nil {
		return errors.New("the api response did not contain a download url")
	}

	zipPath, err := t.downloader.DownloadZip(*url)
	if err != nil {
		return err
	}

	_, err = util.Unzip(zipPath, dir)
	return err
}

func (t *tukUISource) Close() error {
	return t.downloader.Close()
}
