// +build !windows

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func isHidden(t *testing.T, path string) {
	assert.True(t, strings.HasPrefix(filepath.Base(path), "."))
}

func setFileReadOnly(t *testing.T, path string) {
	err := os.Chmod(path, os.FileMode(0400))
	assert.NoError(t, err)
}

func setFileWriteable(t *testing.T, path string) {
	err := os.Chmod(path, os.FileMode(0600))
	assert.NoError(t, err)
}

func Test_HideFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "hidden-files-unix")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	tests := []struct {
		path          func() string
		returnedPath  string
		expectedError bool
		teardown      func(path string)
	}{
		{
			path: func() string {
				f := filepath.Join(dir, "test1")
				err := ioutil.WriteFile(f, []byte{}, os.FileMode(0666))
				assert.NoError(t, err)
				return f
			},
			returnedPath:  filepath.Join(dir, ".test1"),
			expectedError: false,
			teardown: func(path string) {
				os.Remove(path)
			},
		},
		{
			path: func() string {
				f := filepath.Join(dir, ".test2")
				err := ioutil.WriteFile(f, []byte{}, os.FileMode(0666))
				assert.NoError(t, err)
				return f
			},
			returnedPath:  filepath.Join(dir, ".test2"),
			expectedError: false,
			teardown: func(path string) {
				os.Remove(path)
			},
		},
		{
			path: func() string {
				return dir
			},
			returnedPath:  "",
			expectedError: true,
			teardown:      func(path string) {},
		},
		{
			path: func() string {
				return ""
			},
			returnedPath:  "",
			expectedError: true,
			teardown:      func(path string) {},
		},
		{
			path: func() string {
				return "."
			},
			returnedPath:  "",
			expectedError: true,
			teardown:      func(path string) {},
		},
		{
			path: func() string {
				return "fake.file"
			},
			returnedPath:  "",
			expectedError: true,
			teardown:      func(path string) {},
		},
	}

	for _, tt := range tests {
		path := tt.path()

		actual, err := HideFile(path)

		if tt.expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.returnedPath, actual)
			isHidden(t, actual)
		}

		tt.teardown(path)
	}
}
