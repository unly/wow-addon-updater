package sources

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/go-tukui"
	"github.com/unly/wow-addon-updater/updater/sources/mocks"
	"github.com/unly/wow-addon-updater/util/tests"
)

type addonTest struct {
	source        *tukUISource
	addonURL      string
	want          tukui.Addon
	errorExpected bool
	teardown      tests.TearDown
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
		teardown      tests.TearDown
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
		teardown      tests.TearDown
	}

	testMatrix := getUIAddonURLs(t)
	testMatrix = append(testMatrix, getIDAddonURLs(t)...)

	testMatrix = append(testMatrix, func() *addonTest {
		s := newTukUISource()
		return &addonTest{
			source:        s,
			addonURL:      "example.com",
			errorExpected: true,
			teardown:      tests.DeleteDir(t, s.tempDir),
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
		teardown      tests.TearDown
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
		teardown      tests.TearDown
	}

	tests := []func() *test{
		func() *test {
			s := newTukUISource()
			return &test{
				source:        s,
				addonURL:      "example.org",
				dir:           "",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *test {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			resp := &http.Response{
				StatusCode: http.StatusOK,
			}
			m.On("GetTukUI").Return(tukui.Addon{}, resp, nil)
			s.api = m
			return &test{
				source:        s,
				addonURL:      "https://www.tukui.org/download.php?ui=tukui",
				dir:           "",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *test {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)

			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			addon := tukui.Addon{
				URL: stringPtr(server.URL + "/download/addon"),
			}
			resp := &http.Response{
				StatusCode: http.StatusOK,
			}
			m.On("GetClassicAddon", 1).Return(addon, resp, nil)
			s.api = m
			mux.HandleFunc("/download/addon", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				w.WriteHeader(http.StatusInternalServerError)
			})
			teardown := func() {
				server.Close()
				tests.DeleteDir(t, s.tempDir)()
			}
			return &test{
				source:        s,
				addonURL:      "https://www.tukui.org/classic-addons.php?id=1",
				dir:           "",
				errorExpected: true,
				teardown:      teardown,
			}
		},
		func() *test {
			mux := http.NewServeMux()
			server := httptest.NewServer(mux)
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			addon := tukui.Addon{
				URL: stringPtr(server.URL + "/download/addon"),
			}
			resp := &http.Response{
				StatusCode: http.StatusOK,
			}
			m.On("GetClassicAddon", 1).Return(addon, resp, nil)
			s.api = m
			mux.HandleFunc("/download/addon", func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, http.MethodGet, r.Method)
				content, err := ioutil.ReadFile(filepath.Join("_tests", "archive1.zip"))
				assert.NoError(t, err)
				w.Write(content)
				w.WriteHeader(http.StatusOK)
			})
			dir := tests.TempDir(t)
			teardown := func() {
				server.Close()
				tests.DeleteDir(t, s.tempDir)()
				tests.DeleteDir(t, dir)()
			}
			return &test{
				source:        s,
				addonURL:      "https://www.tukui.org/classic-addons.php?id=1",
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
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "tukui.org/abc",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
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
			s.api = m
			return &addonTest{
				source:        s,
				addonURL:      "ui=tukui",
				errorExpected: false,
				want:          addon,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			resp := &http.Response{
				StatusCode: http.StatusInternalServerError,
			}
			m.On("GetTukUI").Return(tukui.Addon{}, resp, nil)
			s.api = m
			return &addonTest{
				source:        s,
				addonURL:      "ui=tukui",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
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
			s.api = m
			return &addonTest{
				source:        s,
				addonURL:      "ui=elvui",
				errorExpected: false,
				want:          addon,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			resp := &http.Response{
				StatusCode: http.StatusInternalServerError,
			}
			m.On("GetElvUI").Return(tukui.Addon{}, resp, nil)
			s.api = m
			return &addonTest{
				source:        s,
				addonURL:      "ui=elvui",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "ui=unsupported",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
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
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			return &addonTest{
				source:        s,
				addonURL:      "tukui.org/classic-addons.php?id=abc",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
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
			m.On("GetClassicAddon", 1).Return(addon, resp, nil)
			s.api = m
			return &addonTest{
				source:        s,
				addonURL:      "tukui.org/classic-addons.php?id=1",
				want:          addon,
				errorExpected: false,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			resp := &http.Response{
				StatusCode: http.StatusBadRequest,
			}
			m.On("GetClassicAddon", 1).Return(tukui.Addon{}, resp, nil)
			s.api = m
			return &addonTest{
				source:        s,
				addonURL:      "tukui.org/classic-addons.php?id=1",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
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
			m.On("GetRetailAddon", 2).Return(addon, resp, nil)
			s.api = m
			return &addonTest{
				source:        s,
				addonURL:      "tukui.org/addons.php?id=2",
				want:          addon,
				errorExpected: false,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
		func() *addonTest {
			s := newTukUISource()
			m := &mocks.MockTukUIAPI{}
			resp := &http.Response{
				StatusCode: http.StatusBadRequest,
			}
			m.On("GetRetailAddon", 2).Return(tukui.Addon{}, resp, nil)
			s.api = m
			return &addonTest{
				source:        s,
				addonURL:      "tukui.org/addons.php?id=2",
				errorExpected: true,
				teardown:      tests.DeleteDir(t, s.tempDir),
			}
		},
	}
}
