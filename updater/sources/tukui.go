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
}

func NewTukUiSource() *TukUiSource {
	return &TukUiSource{
		source:  newSource(`(https?://)?(www\.)?tukui\.org/(classic-)?addons\.php\?id=[0-9]+`, "tukui"),
		client:  tukui.NewClient(nil),
		idRegex: regexp.MustCompile(`id=[0-9]+`),
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
	var addon tukui.Addon

	idString := t.idRegex.FindString(url)
	idString = idString[3:]
	id, err := strconv.Atoi(idString)
	if err != nil {
		return addon, err
	}

	var resp *http.Response
	if strings.Contains(url, "classic-") {
		addon, resp, err = t.client.ClassicAddons.GetAddon(id)
	} else {
		addon, resp, err = t.client.RetailAddons.GetAddon(id)
	}
	if err != nil {
		return addon, err
	}
	if !(resp.StatusCode >= 200 && resp.StatusCode < 300) {
		return addon, fmt.Errorf("failed to get latest version. error code: %s", resp.Status)
	}

	return addon, nil
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
