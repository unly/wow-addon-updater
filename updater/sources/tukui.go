package sources

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/unly/go-tukui"
	"github.com/unly/wow-addon-updater/updater"
	"github.com/unly/wow-addon-updater/util"
)

type tukUISource struct {
	*source
	classic tukuiAPI
	retail  tukuiAPI
	idRegex *regexp.Regexp
	uiRegex *regexp.Regexp
}

//go:generate go run github.com/vektra/mockery/v2 --case=underscore  --name=tukuiAPI --structname=MockTukUIAPI

type tukuiAPI interface {
	tukui.AddonClient
}

// NewTukUISource returns a pointer to a newly created TukUISource.
func NewTukUISource() updater.UpdateSource {
	return newTukUISource()
}

func newTukUISource() *tukUISource {
	client := tukui.NewClient(nil)

	return &tukUISource{
		source:  newSource(regexp.MustCompile(`^(https?://)?(www\.)?tukui\.org/((classic-(tbc-)?)?addons\.php\?id=[0-9]+)|(download\.php\?ui=(tukui|elvui))$`)),
		classic: client.ClassicAddons,
		retail:  client.RetailAddons,
		idRegex: regexp.MustCompile(`id=[0-9]+`),
		uiRegex: regexp.MustCompile(`ui=.+`),
	}
}

// GetLatestVersion returns the latest version for the given addon URL
func (t *tukUISource) GetLatestVersion(addonURL string) (string, error) {
	tukuiAddon, err := t.getAddon(addonURL)
	if err != nil {
		return "", err
	}

	version := tukuiAddon.Version
	if version == nil {
		return "", fmt.Errorf("the api response did not contain a version")
	}

	return *version, nil
}

func (t *tukUISource) getAddon(url string) (tukui.Addon, error) {
	// regular addon parameter queries
	if t.idRegex.FindString(url) != "" {
		return t.getRegularAddon(url)
	}

	// tukui and elvui queries
	if t.uiRegex.FindString(url) != "" {
		return t.getUIAddon(url)
	}

	return tukui.Addon{}, fmt.Errorf("tukui.org url %s is not supported", url)
}

func (t *tukUISource) getUIAddon(url string) (tukui.Addon, error) {
	uiRunes := []rune(t.uiRegex.FindString(url))
	if len(uiRunes) < 3 {
		return tukui.Addon{}, fmt.Errorf("failed to extract the ui= parameter from %s. no ui found", url)
	}

	switch string(uiRunes[3:]) {
	case "tukui":
		tukui, resp, err := t.retail.GetTukUI()
		return tukui, checkHTTPResponse(resp, err)
	case "elvui":
		elvui, resp, err := t.retail.GetElvUI()
		return elvui, checkHTTPResponse(resp, err)
	default:
		return tukui.Addon{}, fmt.Errorf("given tukui.org ui addon link %s is not supported", url)
	}
}

func (t *tukUISource) getRegularAddon(url string) (tukui.Addon, error) {
	addon := tukui.Addon{}

	doc, err := getHTMLPage(url)
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
		return fmt.Errorf("the api response did not contain a download url")
	}

	zipPath, err := t.downloadZip(*url)
	if err != nil {
		return err
	}

	_, err = util.Unzip(zipPath, dir)
	if err != nil {
		return err
	}

	return nil
}
