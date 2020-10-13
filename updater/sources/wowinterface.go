package sources

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/unly/wow-addon-updater/util"
)

// WoWInterfaceSource is the source for addons and UIs hosted on wowinterface.com
type WoWInterfaceSource struct {
	*source
}

// NewWoWInterfaceSource returns a pointer to a newly created WoWInterfaceSource.
func NewWoWInterfaceSource() *WoWInterfaceSource {
	return &WoWInterfaceSource{
		source: newSource(regexp.MustCompile(`(https?://)?(www\.)?wowinterface\.com/downloads/info.+\.html`), "wowinterface"),
	}
}

// GetLatestVersion returns the latest version for the given addon URL
func (w *WoWInterfaceSource) GetLatestVersion(addonURL string) (string, error) {
	doc, err := getHTMLPage(addonURL)
	if err != nil {
		return "", err
	}

	s := doc.Find("#version").Text()
	if !strings.HasPrefix(s, "Version: ") {
		return "", fmt.Errorf("failed to find a version tag for: %s", addonURL)

	}

	return s[9:], nil
}

// DownloadAddon downloads and unzip the addon from the given URL to the given directory
func (w *WoWInterfaceSource) DownloadAddon(addonURL, dir string) error {
	elems := strings.Split(addonURL, "/")
	if len(elems) == 0 {
		return fmt.Errorf("no path to extract from: %s", addonURL)
	}

	name := elems[len(elems)-1]
	name = name[4 : len(name)-5]

	doc, err := getHTMLPage(fmt.Sprintf("https://www.wowinterface.com/downloads/download%s", name))
	if err != nil {
		return err
	}

	link, available := doc.Find(".manuallink > a").Attr("href")
	if !available {
		return fmt.Errorf("failed to find download link for: %s", addonURL)
	}

	zipPath, err := w.downloadZip(link)
	if err != nil {
		return err
	}

	_, err = util.Unzip(zipPath, dir)
	if err != nil {
		return err
	}

	return nil
}

func getHTMLPage(url string) (*goquery.Document, error) {
	resp, err := http.Get(url)
	if err := checkHTTPResponse(resp, err); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return goquery.NewDocumentFromReader(resp.Body)
}
