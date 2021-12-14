//go:build !windows

package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func TestHideFile(t *testing.T) {
	t.Run("hide sample file", func(t *testing.T) {
		dir := helpers.TempDir(t)
		defer helpers.DeleteDir(t, dir)
		f := filepath.Join(dir, "test1")
		err := os.WriteFile(f, []byte{}, os.FileMode(0666))
		if err != nil {
			assert.FailNow(t, "failed to write to test file", err)
		}

		path, err := HideFile(f)

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, ".test1"), path)
	})
	t.Run("hide already hidden file", func(t *testing.T) {
		dir := helpers.TempDir(t)
		defer helpers.DeleteDir(t, dir)
		f := filepath.Join(dir, ".test2")
		err := os.WriteFile(f, []byte{}, os.FileMode(0666))
		if err != nil {
			assert.FailNow(t, "failed to write to test file", err)
		}

		path, err := HideFile(f)

		assert.NoError(t, err)
		assert.Equal(t, filepath.Join(dir, ".test2"), path)
	})
	t.Run("empty path", func(t *testing.T) {
		_, err := HideFile("")

		assert.Error(t, err)
	})
	t.Run("current dir", func(t *testing.T) {
		_, err := HideFile(".")

		assert.Error(t, err)
	})
	t.Run("not existing file", func(t *testing.T) {
		_, err := HideFile("fake.file")

		assert.Error(t, err)
	})
}

func TestIsHiddenFilePath(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{
			path: "",
			want: false,
		},
		{
			path: ".",
			want: false,
		},
		{
			path: ".file",
			want: true,
		},
		{
			path: "/.",
			want: false,
		},
		{
			path: "/",
			want: false,
		},
		{
			path: "/a/.file",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			actual := IsHiddenFilePath(tt.path)

			assert.Equal(t, tt.want, actual)
		})
	}
}

func TestWriteToHiddenFile_Unix(t *testing.T) {
	err := WriteToHiddenFile("dir/test.file", []byte("hello world"), os.FileMode(0666))
	assert.Error(t, err, "WriteToHiddenFile() returned no error")
}
