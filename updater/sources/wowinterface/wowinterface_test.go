package wowinterface

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

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
	source, err := New(nil)
	if err != nil {
		t.FailNow()
	}

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
		t.Run(tt.addonURL, func(t *testing.T) {
			regex := source.GetURLRegex()
			actual := regex.MatchString(tt.addonURL)

			assert.Equal(t, tt.want, actual)
		})
	}
}

func getWoWInterfacePage(version, downloadURL string) string {
	return fmt.Sprintf(wowinterfaceAddonPage, "Version: "+version, downloadURL)
}

func newWoWInterfaceSource(t *testing.T, client *http.Client) *source {
	t.Helper()
	s, err := New(client)
	if err != nil {
		t.FailNow()
	}

	res, ok := s.(*source)
	if !ok {
		t.FailNow()
	}
	return res
}

func Test_GetLatestVersion_WoWInterface(t *testing.T) {
	type getlatestVersionTest struct {
		name          string
		source        *source
		addonURL      string
		want          string
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *getlatestVersionTest{
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				_, _ = rw.Write([]byte(getWoWInterfacePage("1.2.3", "")))
			})
			server := httptest.NewServer(mux)
			s := newWoWInterfaceSource(t, nil)

			return &getlatestVersionTest{
				name:          "example addon",
				source:        s,
				addonURL:      server.URL + "/addon",
				want:          "1.2.3",
				errorExpected: false,
				teardown: func() {
					_ = s.Close()
					server.Close()
				},
			}
		},
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				_, _ = rw.Write([]byte(getWoWInterfacePage("", "")))
			})
			server := httptest.NewServer(mux)
			s := newWoWInterfaceSource(t, nil)

			return &getlatestVersionTest{
				name:          "empty version",
				source:        s,
				addonURL:      server.URL + "/addon",
				want:          "",
				errorExpected: false,
				teardown: func() {
					_ = s.Close()
					server.Close()
				},
			}
		},
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusInternalServerError)
			})
			server := httptest.NewServer(mux)
			s := newWoWInterfaceSource(t, nil)

			return &getlatestVersionTest{
				name:          "internal server error",
				source:        s,
				addonURL:      server.URL + "/addon",
				want:          "",
				errorExpected: true,
				teardown: func() {
					_ = s.Close()
					server.Close()
				},
			}
		},
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				_, _ = rw.Write([]byte(fmt.Sprintf(wowinterfaceAddonPage, "1.2.3", "")))
			})
			server := httptest.NewServer(mux)
			s := newWoWInterfaceSource(t, nil)

			return &getlatestVersionTest{
				name:          "invalid response",
				source:        s,
				addonURL:      server.URL + "/addon",
				want:          "",
				errorExpected: true,
				teardown: func() {
					_ = s.Close()
					server.Close()
				},
			}
		},
		func() *getlatestVersionTest {
			mux := http.NewServeMux()
			mux.HandleFunc("/addon", func(rw http.ResponseWriter, r *http.Request) {
				_, _ = rw.Write([]byte(fmt.Sprintf(wowinterfaceAddonPageNoVersion, "")))
			})
			server := httptest.NewServer(mux)
			s := newWoWInterfaceSource(t, nil)

			return &getlatestVersionTest{
				name:          "no version",
				source:        s,
				addonURL:      server.URL + "/addon",
				want:          "",
				errorExpected: true,
				teardown: func() {
					_ = s.Close()
					server.Close()
				},
			}
		},
	}

	for _, fn := range tests {
		tt := fn()
		t.Run(tt.name, func(t *testing.T) {
			defer tt.teardown()
			actual, err := tt.source.GetLatestVersion(tt.addonURL)

			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, actual)
			}
		})
	}
}

func Test_DownloadAddon_WoWInterface(t *testing.T) {
	type downloadAddonTest struct {
		name          string
		source        *source
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
				content, err := os.ReadFile(filepath.Join("..", "_tests", "archive1.zip"))
				assert.NoError(t, err)
				_, _ = w.Write(content)
			})
			mux.HandleFunc("/downloads/downloadaddon", func(rw http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				_, _ = rw.Write([]byte(fmt.Sprintf(wowinterfaceDownloadPage, server.URL+"/download/addon")))
			})
			dir := helpers.TempDir(t)
			source := newWoWInterfaceSource(t, nil)
			source.baseURL = server.URL
			teardown := func() {
				_ = source.Close()
				server.Close()
			}

			return &downloadAddonTest{
				name:          "download sample file",
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
				_, _ = rw.Write([]byte("Hello World"))
			})
			dir := helpers.TempDir(t)
			source := newWoWInterfaceSource(t, nil)
			source.baseURL = server.URL
			teardown := func() {
				_ = source.Close()
				server.Close()
			}

			return &downloadAddonTest{
				name:          "invalid web payload",
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
			source := newWoWInterfaceSource(t, nil)
			source.baseURL = server.URL
			teardown := func() {
				_ = source.Close()
				server.Close()
			}

			return &downloadAddonTest{
				name:          "internal server error",
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
				_, _ = rw.Write([]byte(fmt.Sprintf(wowinterfaceDownloadPage, "not-existing")))
			})
			dir := helpers.TempDir(t)
			source := newWoWInterfaceSource(t, nil)
			source.baseURL = server.URL
			teardown := func() {
				_ = source.Close()
				server.Close()
			}

			return &downloadAddonTest{
				name:          "not existing addon",
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
		t.Run(tt.name, func(t *testing.T) {
			defer tt.teardown()
			err := tt.source.DownloadAddon(tt.addonURL, tt.dir)

			if tt.errorExpected {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.DirExists(t, tt.outputDir)
			}
		})
	}
}
