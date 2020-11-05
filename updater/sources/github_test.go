package sources

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/unly/wow-addon-updater/updater/sources/mocks"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func Test_getGetLatestRelease(t *testing.T) {
	tests := []struct {
		setup         func() *githubSource
		addonURL      string
		want          *github.RepositoryRelease
		errorExpected bool
	}{
		{
			setup: func() *githubSource {
				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(nil, nil, errors.New("i'm an error"))
				source.api = m
				return source
			},
			addonURL:      "https://github.com/owner/addon",
			want:          nil,
			errorExpected: true,
		},
		{
			setup: func() *githubSource {
				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
					},
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(nil, resp, nil)
				source.api = m
				return source
			},
			addonURL:      "https://github.com/owner/addon",
			want:          nil,
			errorExpected: true,
		},
		{
			setup: func() *githubSource {
				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusInternalServerError,
					},
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(nil, resp, nil)
				source.api = m
				return source
			},
			addonURL:      "https://github.com/owner/addon",
			want:          nil,
			errorExpected: true,
		},
		{
			setup: func() *githubSource {
				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{
					TagName: stringPtr("1.2.3"),
					Name:    stringPtr("addon"),
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m
				return source
			},
			addonURL: "https://github.com/owner/addon",
			want: &github.RepositoryRelease{
				TagName: stringPtr("1.2.3"),
				Name:    stringPtr("addon"),
			},
			errorExpected: false,
		},
		{
			setup:         newGitHubSource,
			addonURL:      "https://github.com/owner",
			want:          nil,
			errorExpected: true,
		},
		{
			setup:         newGitHubSource,
			addonURL:      "",
			want:          nil,
			errorExpected: true,
		},
		{
			setup:         newGitHubSource,
			addonURL:      "https://example.com/owner/addon",
			want:          nil,
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		source := tt.setup()
		_, err := source.getLatestRelease(tt.addonURL)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}
	}
}

func Test_getOrgAndRepository(t *testing.T) {
	tests := []struct {
		setup         func() *githubSource
		addonURL      string
		wantOwner     string
		wantRepo      string
		errorExpected bool
	}{
		{
			setup:         newGitHubSource,
			addonURL:      "",
			wantOwner:     "",
			wantRepo:      "",
			errorExpected: true,
		},
		{
			setup:         newGitHubSource,
			addonURL:      "github.com/owner",
			wantOwner:     "",
			wantRepo:      "",
			errorExpected: true,
		},
		{
			setup:         newGitHubSource,
			addonURL:      "github.com/owner/addon",
			wantOwner:     "owner",
			wantRepo:      "addon",
			errorExpected: false,
		},
		{
			setup:         newGitHubSource,
			addonURL:      "github.com/owner/addon/",
			wantOwner:     "owner",
			wantRepo:      "addon",
			errorExpected: false,
		},
		{
			setup:         newGitHubSource,
			addonURL:      "github.com/owner-1/addon",
			wantOwner:     "owner-1",
			wantRepo:      "addon",
			errorExpected: false,
		},
	}

	for _, tt := range tests {
		source := tt.setup()
		actualOwner, actualRepo, err := source.getOrgAndRepository(tt.addonURL)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.wantOwner, actualOwner)
			assert.Equal(t, tt.wantRepo, actualRepo)
		}
	}
}

func Test_GetLatestVersion(t *testing.T) {
	tests := []struct {
		setup         func() *githubSource
		addonURL      string
		version       string
		errorExpected bool
	}{
		{
			setup:         newGitHubSource,
			addonURL:      "github.com/owner",
			version:       "",
			errorExpected: true,
		},
		{
			setup: func() *githubSource {
				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{
					TagName: stringPtr("1.2.3"),
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m
				return source
			},
			addonURL:      "github.com/owner/addon",
			version:       "1.2.3",
			errorExpected: false,
		},
		{
			setup: func() *githubSource {
				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m
				return source
			},
			addonURL:      "github.com/owner/addon",
			version:       "",
			errorExpected: false,
		},
	}

	for _, tt := range tests {
		source := tt.setup()
		actual, err := source.GetLatestVersion(tt.addonURL)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.version, actual)
		}
	}
}

func Test_GetURLRegex(t *testing.T) {
	source := newGitHubSource()

	tests := []struct {
		addonURL string
		want     bool
	}{
		{
			addonURL: "http://github.com/google/go-github",
			want:     true,
		},
		{
			addonURL: "https://github.com/google/go-github",
			want:     true,
		},
		{
			addonURL: "https://github.com/google/go-github/",
			want:     true,
		},
		{
			addonURL: "https://github.com/google/",
			want:     false,
		},
		{
			addonURL: "github.com/google/go-github/",
			want:     true,
		},
		{
			addonURL: "ftp://github.com/google/go-github",
			want:     false,
		},
	}

	for _, tt := range tests {
		regex := source.GetURLRegex()
		actual := regex.MatchString(tt.addonURL)

		assert.Equal(t, tt.want, actual)
	}
}

func Test_DownloadAddon(t *testing.T) {
	type testStruct struct {
		source        *githubSource
		addonURL      string
		outputDir     string
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []struct {
		setup func() testStruct
	}{
		{
			setup: func() testStruct {
				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusBadRequest,
					},
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(nil, resp, nil)
				source.api = m

				return testStruct{
					source:        source,
					addonURL:      "github.com/owner/addon",
					outputDir:     "",
					errorExpected: true,
					teardown:      helpers.DeleteDir(t, source.tempDir),
				}
			},
		},
		{
			setup: func() testStruct {
				mux := http.NewServeMux()
				server := httptest.NewServer(mux)

				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{
					TagName:    stringPtr("1.2.3"),
					ZipballURL: stringPtr(server.URL + "/download/addon"),
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m

				mux.HandleFunc("/download/addon", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					w.WriteHeader(http.StatusInternalServerError)
				})

				dir := helpers.TempDir(t)

				teardown := func() {
					server.Close()
					helpers.DeleteDir(t, dir)()
					helpers.DeleteDir(t, source.tempDir)()
				}

				return testStruct{
					source:        source,
					addonURL:      "github.com/owner/addon",
					outputDir:     dir,
					errorExpected: true,
					teardown:      teardown,
				}
			},
		},
		{
			setup: func() testStruct {
				mux := http.NewServeMux()
				server := httptest.NewServer(mux)

				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{
					TagName:    stringPtr("1.2.3"),
					ZipballURL: stringPtr(server.URL + "/download/addon"),
					Assets: []*github.ReleaseAsset{
						{},
					},
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m

				mux.HandleFunc("/download/addon", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					w.WriteHeader(http.StatusInternalServerError)
				})

				dir := helpers.TempDir(t)

				teardown := func() {
					server.Close()
					helpers.DeleteDir(t, dir)()
					helpers.DeleteDir(t, source.tempDir)()
				}

				return testStruct{
					source:        source,
					addonURL:      "github.com/owner/addon",
					outputDir:     dir,
					errorExpected: true,
					teardown:      teardown,
				}
			},
		},
		{
			setup: func() testStruct {
				mux := http.NewServeMux()
				server := httptest.NewServer(mux)

				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{
					TagName:    stringPtr("1.2.3"),
					ZipballURL: stringPtr(server.URL + "/download/addon"),
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m

				mux.HandleFunc("/download/addon", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					w.WriteHeader(http.StatusInternalServerError)
				})

				dir := helpers.TempDir(t)

				teardown := func() {
					server.Close()
					helpers.DeleteDir(t, dir)()
					helpers.DeleteDir(t, source.tempDir)()
				}

				return testStruct{
					source:        source,
					addonURL:      "github.com/owner/addon",
					outputDir:     dir,
					errorExpected: true,
					teardown:      teardown,
				}
			},
		},
		{
			setup: func() testStruct {
				mux := http.NewServeMux()
				server := httptest.NewServer(mux)

				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{
					TagName:    stringPtr("1.2.3"),
					ZipballURL: stringPtr(server.URL + "/download/addon"),
					Assets: []*github.ReleaseAsset{
						{
							BrowserDownloadURL: stringPtr(server.URL + "/download/asset"),
							ContentType:        stringPtr("application/x-zip-compressed"),
						},
					},
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m

				mux.HandleFunc("/download/asset", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					w.WriteHeader(http.StatusInternalServerError)
				})

				dir := helpers.TempDir(t)

				teardown := func() {
					server.Close()
					helpers.DeleteDir(t, dir)()
					helpers.DeleteDir(t, source.tempDir)()
				}

				return testStruct{
					source:        source,
					addonURL:      "github.com/owner/addon",
					outputDir:     dir,
					errorExpected: true,
					teardown:      teardown,
				}
			},
		},
		{
			setup: func() testStruct {
				mux := http.NewServeMux()
				server := httptest.NewServer(mux)

				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{
					TagName:    stringPtr("1.2.3"),
					ZipballURL: stringPtr(server.URL + "/download/addon"),
					Assets: []*github.ReleaseAsset{
						{
							BrowserDownloadURL: stringPtr(server.URL + "/download/asset"),
							ContentType:        stringPtr("fake"),
						},
					},
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m

				mux.HandleFunc("/download/addon", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					content, err := ioutil.ReadFile(filepath.Join("_tests", "archive1.zip"))
					assert.NoError(t, err)
					w.Write(content)
					w.WriteHeader(http.StatusOK)
				})

				dir := helpers.TempDir(t)

				teardown := func() {
					server.Close()
					helpers.DeleteDir(t, dir)()
					helpers.DeleteDir(t, source.tempDir)()
				}

				return testStruct{
					source:        source,
					addonURL:      "github.com/owner/addon",
					outputDir:     dir,
					errorExpected: false,
					teardown:      teardown,
				}
			},
		},
		{
			setup: func() testStruct {
				mux := http.NewServeMux()
				server := httptest.NewServer(mux)

				source := newGitHubSource()
				m := &mocks.MockGitHubAPI{}
				resp := &github.Response{
					Response: &http.Response{
						StatusCode: http.StatusOK,
					},
				}
				release := &github.RepositoryRelease{
					TagName:    stringPtr("1.2.3"),
					ZipballURL: stringPtr(server.URL + "/download/addon"),
				}
				m.On("GetLatestRelease", mock.Anything, "owner", "addon").Return(release, resp, nil)
				source.api = m

				mux.HandleFunc("/download/addon", func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, http.MethodGet, r.Method)
					w.WriteHeader(http.StatusOK)
				})

				dir := helpers.TempDir(t)

				teardown := func() {
					server.Close()
					helpers.DeleteDir(t, dir)()
					helpers.DeleteDir(t, source.tempDir)()
				}

				return testStruct{
					source:        source,
					addonURL:      "github.com/owner/addon",
					outputDir:     dir,
					errorExpected: true,
					teardown:      teardown,
				}
			},
		},
	}

	for _, tt := range tests {
		test := tt.setup()

		err := test.source.DownloadAddon(test.addonURL, test.outputDir)

		if test.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		test.teardown()
	}
}

func stringPtr(s string) *string {
	return &s
}
