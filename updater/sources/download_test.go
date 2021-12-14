package sources

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func TestClose(t *testing.T) {
	tests := []struct {
		name          string
		source        *source
		errorExpected bool
	}{
		{
			name: "tmp dir",
			source: &source{
				tempDir: helpers.TempDir(t),
			},
		},
		{
			name: "not existing dir",
			source: &source{
				tempDir: "not existing",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.source.Close()

			assert.Equal(t, tt.errorExpected, err != nil)
			assert.NoDirExists(t, tt.source.tempDir)
		})
	}
}

func TestDownloadZip(t *testing.T) {
	t.Run("invalid url", func(t *testing.T) {
		d := source{client: http.DefaultClient}

		_, err := d.DownloadZip("invalid url")

		assert.Error(t, err)
	})
	t.Run("error response", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(http.StatusInternalServerError)
		})
		s := httptest.NewServer(mux)
		defer s.Close()

		d := source{client: http.DefaultClient}

		_, err := d.DownloadZip(s.URL)

		assert.Error(t, err)
	})
	t.Run("no local zip file", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(rw http.ResponseWriter, _ *http.Request) {
			rw.WriteHeader(http.StatusOK)
		})
		s := httptest.NewServer(mux)
		defer s.Close()

		d := source{client: http.DefaultClient, tempDir: "not existing"}

		_, err := d.DownloadZip(s.URL)

		assert.Error(t, err)
	})
	t.Run("download zip file", func(t *testing.T) {
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(rw http.ResponseWriter, _ *http.Request) {
			content, err := os.ReadFile(filepath.Join("_tests", "archive1.zip"))
			if err != nil {
				assert.FailNow(t, "failed to read in dummy zip file", err)
			}
			rw.Write(content)
		})
		s := httptest.NewServer(mux)
		defer s.Close()
		dir := helpers.TempDir(t)
		d := source{client: http.DefaultClient, tempDir: dir}
		defer d.Close()

		file, err := d.DownloadZip(s.URL)

		assert.NoError(t, err)
		assert.True(t, strings.HasSuffix(file, ".zip"))
		assert.FileExists(t, file)
	})
}
