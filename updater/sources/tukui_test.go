package sources

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/go-tukui"
	"github.com/unly/wow-addon-updater/updater/sources/mocks"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

const (
	tukuiAddonPage string = `
<!DOCTYPE html>
<html lang="en-US">
	<head>
		<title>Tukui</title>
	</head>
	<body class="appear-animate">
	<ul class="nav nav-tabs tpl-tabs animate">
		<li class="addons-single in active">
			<a href="#description" data-toggle="tab">Description</a>
		</li>
		<li class="addons-single">
			<a href="#screenshot" data-toggle="tab">Screenshot</a>
		</li>
		<li class="addons-single ">
			<a href="#changelog" data-toggle="tab">Changelog</a>
		</li>
		<li class="addons-single">
			<a href="#extras" data-toggle="tab">Extras</a>
		</li>
	</ul>
	<div class="tab-pane fade in" id="extras">
		<p class="extras">The latest version of this addon is <b class="VIP">%s</b> and was uploaded on <b class="VIP">Oct 27, 2020</b> at <b class="VIP">02:17</b>.</p>
		<p class="extras">This file was last downloaded on <b class="VIP">Dec 09, 2020</b> at <b class="VIP">21:48</b> and has been downloaded <b class="VIP">1572354</b> times.</p>
	</div>
	</body>
</html>
	`
)

type addonTest struct {
	source        *tukUISource
	addonURL      string
	want          tukui.Addon
	errorExpected bool
	teardown      helpers.TearDown
}

func Test_getUIAddon(t *testing.T) {

	tests := getUIAddonURLs(t)

	for _, fn := range tests {
		tt := fn()

		actual, err := tt.source.getUIAddon(tt.addonURL)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		}

		tt.teardown()
	}
}

func Test_getRegularAddon(t *testing.T) {
	type test struct {
		source        *tukUISource
		addonURL      string
		want          tukui.Addon
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := getIDAddonURLs(t)

	for _, fn := range tests {
		tt := fn()

		actual, err := tt.source.getRegularAddon(tt.addonURL)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		}

		tt.teardown()
	}
}

func Test_getAddon(t *testing.T) {
	type test struct {
		source        *tukUISource
		addonURL      string
		want          tukui.Addon
		errorExpected bool
		teardown      helpers.TearDown
	}

	testMatrix := getUIAddonURLs(t)
	testMatrix = append(testMatrix, getIDAddonURLs(t)...)

	testMatrix = append(testMatrix, func() *addonTest {
		s := newTukUISource()
		return &addonTest{
			source:        s,
			addonURL:      "example.com",
			errorExpected: true,
			teardown:      helpers.DeleteDir(t, s.tempDir),
		}
	})

	for _, fn := range testMatrix {
		tt := fn()

		actual, err := tt.source.getAddon(tt.addonURL)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, actual)
		}

		tt.teardown()
	}
}

func Test_GetLatestVersion_TukUI(t *testing.T) {
	type test struct {
		source        *tukUISource
		addonURL      string
		want          string
		errorExpected bool
		teardown      helpers.TearDown
	}

	addonTests := getUIAddonURLs(t)
	addonTests = append(addonTests, getIDAddonURLs(t)...)

	tests := make([]test, len(addonTests))
	for i, fn := range addonTests {
		x := fn()
		want := ""
		if x.want.Version != nil {
			want = *x.want.Version
		}
		tests[i] = test{
			source:        x.source,
			addonURL:      x.addonURL,
			want:          want,
			errorExpected: x.errorExpected,
			teardown:      x.teardown,
		}
	}

	for _, tt := range tests {
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

func Test_DownloadAddon_TukUI(t *testing.T) {
	type test struct {
		source        *tukUISource
		addonURL      string
		dir           string
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *test{
		func() *test {
			s := newTukUISource()
			return &test{
				source:        s,
				addonURL:      "example.org",
				dir:           "",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *test {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			resp := &http.Response{
				StatusCode: http.StatusOK,
			}
			m.On("GetTukUI").Return(tukui.Addon{}, resp, nil)
			s.retail = m
			return &test{
				source:        s,
				addonURL:      "https://www.tukui.org/download.php?ui=tukui",
				dir:           "",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *test {
			mux := http.NewServeMux()
			mux.HandleFunc("/classic-addons.php", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				v := r.URL.Query()
				if v.Get("id") == "1" {
					w.Write([]byte(fmt.Sprintf(tukuiAddonPage, "1.2.3")))
					w.WriteHeader(http.StatusOK)
				} else if v.Get("download") == "1" {
					w.WriteHeader(http.StatusInternalServerError)
				}
			})
			server := httptest.NewServer(mux)
			s := newTukUISource()
			teardown := func() {
				server.Close()
				helpers.DeleteDir(t, s.tempDir)()
			}
			return &test{
				source:        s,
				addonURL:      server.URL + "/classic-addons.php?id=1",
				dir:           "",
				errorExpected: true,
				teardown:      teardown,
			}
		},
		func() *test {
			mux := http.NewServeMux()
			mux.HandleFunc("/classic-addons.php", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				v := r.URL.Query()
				if v.Get("id") == "1" {
					w.Write([]byte(fmt.Sprintf(tukuiAddonPage, "1.2.3")))
					w.WriteHeader(http.StatusOK)
				} else if v.Get("download") == "1" {
					content, err := ioutil.ReadFile(filepath.Join("_tests", "archive1.zip"))
					assert.NoError(t, err)
					w.Write(content)
					w.WriteHeader(http.StatusOK)
				}
			})
			server := httptest.NewServer(mux)
			dir := helpers.TempDir(t)
			s := newTukUISource()
			teardown := func() {
				server.Close()
				helpers.DeleteDir(t, s.tempDir)()
				helpers.DeleteDir(t, dir)()
			}
			return &test{
				source:        s,
				addonURL:      server.URL + "/classic-addons.php?id=1",
				dir:           dir,
				errorExpected: false,
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
		}

		tt.teardown()
	}
}

func getUIAddonURLs(t *testing.T) []func() *addonTest {
	return []func() *addonTest{
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "example.com",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "tukui.org/abc",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			addon := tukui.Addon{
				Version: stringPtr("1.2.3"),
			}
			resp := &http.Response{
				StatusCode: http.StatusOK,
			}
			m.On("GetTukUI").Return(addon, resp, nil)
			s.retail = m
			return &addonTest{
				source:        s,
				addonURL:      "ui=tukui",
				errorExpected: false,
				want:          addon,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			resp := &http.Response{
				StatusCode: http.StatusInternalServerError,
			}
			m.On("GetTukUI").Return(tukui.Addon{}, resp, nil)
			s.retail = m
			return &addonTest{
				source:        s,
				addonURL:      "ui=tukui",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			addon := tukui.Addon{
				Version: stringPtr("1.2.3"),
			}
			resp := &http.Response{
				StatusCode: http.StatusOK,
			}
			m.On("GetElvUI").Return(addon, resp, nil)
			s.retail = m
			return &addonTest{
				source:        s,
				addonURL:      "ui=elvui",
				errorExpected: false,
				want:          addon,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			resp := &http.Response{
				StatusCode: http.StatusInternalServerError,
			}
			m.On("GetElvUI").Return(tukui.Addon{}, resp, nil)
			s.retail = m
			return &addonTest{
				source:        s,
				addonURL:      "ui=elvui",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "ui=unsupported",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
	}
}

func getIDAddonURLs(t *testing.T) []func() *addonTest {
	return []func() *addonTest{
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "tukui.org/classic-addons.php?id=abc",
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			website := fmt.Sprintf(tukuiAddonPage, "1.2.3")
			mux := http.NewServeMux()
			mux.HandleFunc("/classic-addons.php", func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte(website))
				rw.WriteHeader(http.StatusOK)
			})
			server := httptest.NewServer(mux)
			addon := tukui.Addon{
				Version: stringPtr("1.2.3"),
				URL:     stringPtr(server.URL + "/classic-addons.php?download=1"),
			}
			return &addonTest{
				source:        s,
				addonURL:      server.URL + "/classic-addons.php?id=1",
				want:          addon,
				errorExpected: false,
				teardown: func() {
					helpers.DeleteDir(t, s.tempDir)
					server.Close()
				},
			}
		},
		func() *addonTest {
			s := newTukUISource()
			mux := http.NewServeMux()
			mux.HandleFunc("/addons.php", func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte("not a website"))
				rw.WriteHeader(http.StatusOK)
			})
			server := httptest.NewServer(mux)
			return &addonTest{
				source:        s,
				addonURL:      server.URL + "/classic-addons.php?id=1",
				errorExpected: true,
				teardown: func() {
					helpers.DeleteDir(t, s.tempDir)
					server.Close()
				},
			}
		},
		func() *addonTest {
			s := newTukUISource()
			website := fmt.Sprintf(tukuiAddonPage, "1.2.3")
			mux := http.NewServeMux()
			mux.HandleFunc("/addons.php", func(rw http.ResponseWriter, r *http.Request) {
				rw.Write([]byte(website))
				rw.WriteHeader(http.StatusOK)
			})
			server := httptest.NewServer(mux)
			addon := tukui.Addon{
				Version: stringPtr("1.2.3"),
				URL:     stringPtr(server.URL + "/addons.php?download=2"),
			}
			return &addonTest{
				source:        s,
				addonURL:      server.URL + "/addons.php?id=2",
				want:          addon,
				errorExpected: false,
				teardown: func() {
					helpers.DeleteDir(t, s.tempDir)
					server.Close()
				},
			}
		},
		func() *addonTest {
			s := newTukUISource()
			mux := http.NewServeMux()
			mux.HandleFunc("/addons.php", func(rw http.ResponseWriter, r *http.Request) {
				rw.WriteHeader(http.StatusBadRequest)
			})
			server := httptest.NewServer(mux)
			return &addonTest{
				source:        s,
				addonURL:      server.URL + "/addons.php?id=2",
				errorExpected: true,
				teardown: func() {
					helpers.DeleteDir(t, s.tempDir)
					server.Close()
				},
			}
		},
	}
}
