// +build windows

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

func isHidden(t *testing.T, path string) {
	filenameW, err := syscall.UTF16PtrFromString(path)
	assert.NoError(t, err)

	attrs, err := syscall.GetFileAttributes(filenameW)
	assert.NoError(t, err)

	assert.True(t, attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0)
}

func Test_HideFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "hidden-files-windows")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	tests := []struct {
		path          func() string
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
				filenameW, err := syscall.UTF16PtrFromString(f)
				assert.NoError(t, err)
				err = syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
				assert.NoError(t, err)
				return f
			},
			expectedError: false,
			teardown: func(path string) {
				os.Remove(path)
			},
		},
		{
			path: func() string {
				return dir
			},
			expectedError: true,
			teardown:      func(path string) {},
		},
		{
			path: func() string {
				return ""
			},
			expectedError: true,
			teardown:      func(path string) {},
		},
		{
			path: func() string {
				return "."
			},
			expectedError: true,
			teardown:      func(path string) {},
		},
		{
			path: func() string {
				return "fake.file"
			},
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
			assert.Equal(t, path, actual)
			isHidden(t, actual)
		}

		tt.teardown(path)
	}
}
