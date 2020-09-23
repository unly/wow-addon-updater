package sources

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/unly/go-tukui"
	"github.com/unly/wow-addon-updater/util"
)

type TukUiSource struct {
	*source
	client  *tukui.Client
	idRegex *regexp.Regexp
	uiRegex *regexp.Regexp
}

func NewTukUiSource() *TukUiSource {
	return &TukUiSource{
		source:  newSource(`(https?://)?(www\.)?tukui\.org/((classic-)?addons\.php\?id=[0-9]+)|(download\.php\?ui=(tukui|elvui))`, "tukui"),
		client:  tukui.NewClient(nil),
		idRegex: regexp.MustCompile(`id=[0-9]+`),
		uiRegex: regexp.MustCompile(`ui=(tukui|elvui)`),
	}
}

func (t *TukUiSource) GetLatestVersion(addon string) (string, error) {
	tukuiAddon, err := t.getAddon(addon)
	if err != nil {
		return "", err
	}

	version := tukuiAddon.Version
	if version == nil {
		return "", fmt.Errorf("the api response did not contain a version")
	}

	return *version, nil
}

func (t *TukUiSource) getAddon(url string) (tukui.Addon, error) {
	// regular addon parameter queries
	if t.idRegex.FindString(url) != "" {
		return t.getRegularAddon(url)
	}

	// tukui and elvui queries
	ui := t.uiRegex.FindString(url)
	if len(ui) < 3 {
		return tukui.Addon{}, fmt.Errorf("failed to match: %s to a tukui.org query", url)
	}

	switch ui[3:] {
	case "tukui":
		tukui, resp, err := t.client.RetailAddons.GetTukUI()
		return tukui, checkHTTPResponse(resp, err)
	case "elvui":
		elvui, resp, err := t.client.RetailAddons.GetElvUI()
		return elvui, checkHTTPResponse(resp, err)
	default:
		return tukui.Addon{}, fmt.Errorf("given tukui.org ui addon link is not supported")
	}
}

func (t *TukUiSource) getRegularAddon(url string) (tukui.Addon, error) {
	var addon tukui.Addon

	idString := t.idRegex.FindString(url)
	if len(idString) < 3 {
		return tukui.Addon{}, fmt.Errorf("failed to match: %s to a tukui.org query", url)
	}

	id, err := strconv.Atoi(idString[3:])
	if err != nil {
		return addon, err
	}

	var resp *http.Response
	if strings.Contains(url, "classic-") {
		addon, resp, err = t.client.ClassicAddons.GetAddon(id)
	} else {
		addon, resp, err = t.client.RetailAddons.GetAddon(id)
	}

	return addon, checkHTTPResponse(resp, err)
}

func (t *TukUiSource) DownloadAddon(addon, dir string) error {
	tukuiAddon, err := t.getAddon(addon)
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
