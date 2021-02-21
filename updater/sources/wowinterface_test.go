package sources

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

const (
	wowinterfaceAddonPage string = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" dir="ltr" lang="en">
<head>
	<title>I am an addon</title>
</head>
<body>
	<div>
		<div id="version">%s</div>
	<div>
	<div id="download">
		<div id="size">(446 Kb)</div>
		<a href="%s" title="WoW Retail">Download</a>
	</div>
</body>
</html>
`
	wowinterfaceAddonPageNoVersion string = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" dir="ltr" lang="en">
<head>
	<title>I am an addon</title>
</head>
<body>
	<div id="download">
		<div id="size">(446 Kb)</div>
		<a href="%s" title="WoW Retail">Download</a>
	</div>
</body>
</html>
`
	wowinterfaceDownloadPage string = `
<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Transitional//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-transitional.dtd">
<html xmlns="http://www.w3.org/1999/xhtml" dir="ltr" lang="en">
<head>
	<title>I am an addon</title>
</head>
<body>
	<div class="manuallink">
		Problems with the download? <a href="%s">Click here</a>.
	</div>
</body>
</html>
`
)

func Test_GetURLRegex_WoWinterface(t *testing.T) {
	source := newWoWInterfaceSource()

	tests := []struct {
		addonURL string
		want     bool
	}{
		{
			addonURL: "https://www.wowinterface.com/downloads/info25118-DejaClassicStats.html",
			want:     true,
		},
		{
			addonURL: "https://www.wowinterface.com/downloads/infoabc.html",
			want:     true,
		},
		{
			addonURL: "https://www.wowinterface.com/downloads/info25118-DejaClassicStats",
			want:     false,
		},
		{
			addonURL: "https://wowinterface.com/downloads/info25118-DejaClassicStats.html",
			want:     true,
		},
		{
			addonURL: "wowinterface.com/downloads/info25118-DejaClassicStats.html",
			want:     true,
		},
		{
			addonURL: "ftp://wowinterface.com/downloads/info25118-DejaClassicStats.html",
			want:     false,
		},
	}

	for _, tt := range tests {
		regex := source.GetURLRegex()
		actual := regex.MatchString(tt.addonURL)

		assert.Equal(t, tt.want, actual)
	}
}

func getWoWInterfacePage(version, downloadURL string) string {
	return fmt.Sprintf(wowinterfaceAddonPage, "Version: "+version, downloadURL)
}

func Test_getHTMLPage(t *testing.T) {
	type getHTMLPageTest struct {
		url           string
		document      *goquery.Document
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *getHTMLPageTest{
		func() *getHTMLPageTest {
			website := getWoWInterfacePage("1.2.3", "")
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(website))
			})
			server := httptest.NewServer(mux)
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(website))
			assert.NoError(t, err)

			return &getHTMLPageTest{
				url:           server.URL + "/addon",
				document:      doc,
				errorExpected: false,
				teardown:      server.Close,
			}
		},
		func() *getHTMLPageTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusBadRequest)
			})
			server := httptest.NewServer(mux)

			return &getHTMLPageTest{
				url:           server.URL + "/addon",
				document:      nil,
				errorExpected: true,
				teardown:      server.Close,
			}
		},
		func() *getHTMLPageTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
			})
			server := httptest.NewServer(mux)
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(""))
			assert.NoError(t, err)

			return &getHTMLPageTest{
				url:           server.URL + "/addon",
				document:      doc,
				errorExpected: false,
				teardown:      server.Close,
			}
		},
		func() *getHTMLPageTest {
			s := "hello world"
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(s))
			})
			server := httptest.NewServer(mux)
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(s))
			assert.NoError(t, err)

			return &getHTMLPageTest{
				url:           server.URL + "/addon",
				document:      doc,
				errorExpected: false,
				teardown:      server.Close,
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		actual, err := getHTMLPage(tt.url)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.document, actual)
		}

		tt.teardown()
	}
}

func Test_GetLatestVersion_WoWInterface(t *testing.T) {
	type getlatestVersionTest struct {
		source        *woWInterfaceSource
		addonURL      string
		want          string
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *getlatestVersionTest{
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(getWoWInterfacePage("1.2.3", "")))
			})
			server := httptest.NewServer(mux)

			return &getlatestVersionTest{
				source:        newWoWInterfaceSource(),
				addonURL:      server.URL + "/addon",
				want:          "1.2.3",
				errorExpected: false,
				teardown:      server.Close,
			}
		},
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(getWoWInterfacePage("", "")))
			})
			server := httptest.NewServer(mux)

			return &getlatestVersionTest{
				source:        newWoWInterfaceSource(),
				addonURL:      server.URL + "/addon",
				want:          "",
				errorExpected: false,
				teardown:      server.Close,
			}
		},
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusInternalServerError)
			})
			server := httptest.NewServer(mux)

			return &getlatestVersionTest{
				source:        newWoWInterfaceSource(),
				addonURL:      server.URL + "/addon",
				want:          "",
				errorExpected: true,
				teardown:      server.Close,
			}
		},
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(fmt.Sprintf(wowinterfaceAddonPage, "1.2.3", "")))
			})
			server := httptest.NewServer(mux)

			return &getlatestVersionTest{
				source:        newWoWInterfaceSource(),
				addonURL:      server.URL + "/addon",
				want:          "",
				errorExpected: true,
				teardown:      server.Close,
			}
		},
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(fmt.Sprintf(wowinterfaceAddonPageNoVersion, "")))
			})
			server := httptest.NewServer(mux)

			return &getlatestVersionTest{
				source:        newWoWInterfaceSource(),
				addonURL:      server.URL + "/addon",
				want:          "",
				errorExpected: true,
				teardown:      server.Close,
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		actual, err := tt.source.GetLatestVersion(tt.addonURL)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		}

		tt.teardown()
	}
}

func Test_DownloadAddon_WoWInterface(t *testing.T) {
	type downloadAddonTest struct {
		source        *woWInterfaceSource
		addonURL      string
		dir           string
		outputDir     string
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *downloadAddonTest{
		func() *downloadAddonTest {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			mux.HandleFunc("/download/addon", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				content, err := os.ReadFile(filepath.Join("_tests", "archive1.zip"))
				assert.NoError(t, err)
				w.WriteHeader(http.StatusOK)
				w.Write(content)
			})
			mux.HandleFunc("/downloads/downloadaddon", func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				rw.Write([]byte(fmt.Sprintf(wowinterfaceDownloadPage, server.URL+"/download/addon")))
				rw.WriteHeader(http.StatusOK)
			})
			dir := helpers.TempDir(t)
			source := newWoWInterfaceSource()
			source.baseURL = server.URL
			teardown := func() {
				server.Close()
				helpers.DeleteDir(t, dir)
			}

			return &downloadAddonTest{
				source:        source,
				addonURL:      server.URL + "/infoaddon.html",
				dir:           dir,
				outputDir:     dir + "/root",
				errorExpected: false,
				teardown:      teardown,
			}
		},
		func() *downloadAddonTest {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			mux.HandleFunc("/downloads/downloadaddon", func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte("Hello World"))
			})
			dir := helpers.TempDir(t)
			source := newWoWInterfaceSource()
			source.baseURL = server.URL
			teardown := func() {
				server.Close()
				helpers.DeleteDir(t, dir)
			}

			return &downloadAddonTest{
				source:        source,
				addonURL:      server.URL + "/infoaddon.html",
				dir:           dir,
				outputDir:     "",
				errorExpected: true,
				teardown:      teardown,
			}
		},
		func() *downloadAddonTest {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			mux.HandleFunc("/downloads/downloadaddon", func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				rw.WriteHeader(http.StatusInternalServerError)
			})
			dir := helpers.TempDir(t)
			source := newWoWInterfaceSource()
			source.baseURL = server.URL
			teardown := func() {
				server.Close()
				helpers.DeleteDir(t, dir)
			}

			return &downloadAddonTest{
				source:        source,
				addonURL:      server.URL + "/infoaddon.html",
				dir:           dir,
				outputDir:     "",
				errorExpected: true,
				teardown:      teardown,
			}
		},
		func() *downloadAddonTest {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			mux.HandleFunc("/downloads/downloadaddon", func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				rw.WriteHeader(http.StatusOK)
				rw.Write([]byte(fmt.Sprintf(wowinterfaceDownloadPage, "not-existing")))
			})
			dir := helpers.TempDir(t)
			source := newWoWInterfaceSource()
			source.baseURL = server.URL
			teardown := func() {
				server.Close()
				helpers.DeleteDir(t, dir)
			}

			return &downloadAddonTest{
				source:        source,
				addonURL:      server.URL + "/infoaddon.html",
				dir:           dir,
				outputDir:     "",
				errorExpected: true,
				teardown:      teardown,
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		err := tt.source.DownloadAddon(tt.addonURL, tt.dir)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.DirExists(t, tt.outputDir)
		}

		tt.teardown()
	}
}
