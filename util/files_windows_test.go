//go:build windows

package util

import (
	"os"
	"path/filepath"
	"syscall"
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
		assert.Equal(t, filepath.Join(dir, "test1"), path)
	})
	t.Run("hide already hidden file", func(t *testing.T) {
		dir := helpers.TempDir(t)
		defer helpers.DeleteDir(t, dir)
		f := filepath.Join(dir, ".test2")
		err := os.WriteFile(f, []byte{}, os.FileMode(0666))
		if err != nil {
			assert.FailNow(t, "failed to write to test file", err)
		}
		err = setFileAttribute(f, syscall.FILE_ATTRIBUTE_HIDDEN)
		if err != nil {
			assert.FailNow(t, "failed to set hidden file flag", err)
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
	tests := []string{
		"",
		".",
		"file",
		".file",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			actual := IsHiddenFilePath(tt)

			assert.True(t, actual)
		})
	}
}
