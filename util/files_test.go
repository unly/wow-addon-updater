package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func setFileReadOnly(t *testing.T, path string) {
	t.Helper()
	err := os.Chmod(path, os.FileMode(0400))
	if err != nil {
		assert.FailNow(t, "failed to set read only flag", err)
	}
}

func setFileWriteable(t *testing.T, path string) {
	t.Helper()
	err := os.Chmod(path, os.FileMode(0600))
	if err != nil {
		assert.FailNow(t, "failed to set write flag", err)
	}
}

func TestFileExists(t *testing.T) {
	t.Run("empty file", func(t *testing.T) {
		file := helpers.TempFile(t, "", []byte{})
		defer helpers.DeleteFile(t, file)

		got := FileExists(file)

		assert.True(t, got)
	})
	t.Run("directory", func(t *testing.T) {
		dir := helpers.TempDir(t)
		defer helpers.DeleteDir(t, dir)

		got := FileExists(dir)

		assert.False(t, got)
	})
	t.Run("empty string", func(t *testing.T) {
		got := FileExists("")

		assert.False(t, got)
	})
	t.Run("current dir", func(t *testing.T) {
		got := FileExists(".")

		assert.False(t, got)
	})
}

func TestIsHiddenFile(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		hidden, err := IsHiddenFile("")

		assert.NoError(t, err)
		assert.False(t, hidden)
	})
	t.Run("current dir", func(t *testing.T) {
		hidden, err := IsHiddenFile(".")

		assert.NoError(t, err)
		assert.False(t, hidden)
	})
	t.Run("not existing file", func(t *testing.T) {
		hidden, err := IsHiddenFile("fake file")

		assert.NoError(t, err)
		assert.False(t, hidden)
	})
	t.Run("not hidden file", func(t *testing.T) {
		file := helpers.TempFile(t, "", []byte{})
		defer helpers.DeleteFile(t, file)

		hidden, err := IsHiddenFile(file)

		assert.NoError(t, err)
		assert.False(t, hidden)
	})
	t.Run("hidden file", func(t *testing.T) {
		file, err := HideFile(helpers.TempFile(t, "", []byte{}))
		if err != nil {
			assert.FailNow(t, "failed to hide file", err)
		}

		hidden, err := IsHiddenFile(file)

		assert.NoError(t, err)
		assert.True(t, hidden)
	})
}

func TestWriteToHiddenFile(t *testing.T) {
	dir := helpers.TempDir(t)
	defer helpers.DeleteDir(t, dir)

	t.Run("write to new file", func(t *testing.T) {
		file := filepath.Join(dir, ".test")
		defer helpers.DeleteFile(t, file)

		err := WriteToHiddenFile(file, []byte("hello world"), os.FileMode(0666))

		assert.NoError(t, err)
	})
	t.Run("write to existing, hidden file", func(t *testing.T) {
		f := helpers.TempFile(t, dir, []byte{})
		defer helpers.DeleteFile(t, f)
		file, err := HideFile(f)
		if err != nil {
			assert.FailNow(t, "failed to hide existing file", err)
		}

		err = WriteToHiddenFile(file, []byte("hello world"), os.FileMode(0666))

		assert.NoError(t, err)
	})
	t.Run("overwrite existing content", func(t *testing.T) {
		f := helpers.TempFile(t, dir, []byte("old content"))
		defer helpers.DeleteFile(t, f)
		file, err := HideFile(f)
		if err != nil {
			assert.FailNow(t, "failed to hide existing file", err)
		}

		err = WriteToHiddenFile(file, []byte("new content"), os.FileMode(0666))

		assert.NoError(t, err)
		content, err := os.ReadFile(file)
		assert.NoError(t, err)
		assert.Equal(t, []byte("new content"), content)
	})
	t.Run("try to write to read only file", func(t *testing.T) {
		f := helpers.TempFile(t, "", []byte{})
		defer helpers.DeleteFile(t, f)
		file, err := HideFile(f)
		if err != nil {
			assert.FailNow(t, "failed to hide existing file", err)
		}
		setFileReadOnly(t, file)
		defer setFileWriteable(t, file)

		err = WriteToHiddenFile(file, []byte("i should not be there"), os.FileMode(0666))

		assert.Error(t, err)
	})
}

func TestUnzip(t *testing.T) {
	t.Run("not existing src", func(t *testing.T) {
		_, err := Unzip("not-existing", ".")

		assert.Error(t, err)
	})
	t.Run("current dir src", func(t *testing.T) {
		_, err := Unzip(".", ".")

		assert.Error(t, err)
	})
	t.Run("multi file archive", func(t *testing.T) {
		dest := helpers.TempDir(t)
		defer helpers.DeleteDir(t, dest)

		got, err := Unzip(filepath.Join("tests", "archive1.zip"), dest)

		assert.NoError(t, err)
		want := []string{
			filepath.Join(dest, "a.txt"),
			filepath.Join(dest, "b.txt"),
			filepath.Join(dest, "c.txt"),
		}
		assert.Equal(t, want, got)
	})
	t.Run("single file archive", func(t *testing.T) {
		dest := helpers.TempDir(t)
		defer helpers.DeleteDir(t, dest)

		got, err := Unzip(filepath.Join("tests", "archive2.zip"), dest)

		assert.NoError(t, err)
		want := []string{
			filepath.Join(dest, "a.txt"),
		}
		assert.Equal(t, want, got)
	})
	t.Run("empty archive", func(t *testing.T) {
		dest := helpers.TempDir(t)
		defer helpers.DeleteDir(t, dest)

		got, err := Unzip(filepath.Join("tests", "archive3.zip"), dest)

		assert.NoError(t, err)
		assert.Equal(t, []string{}, got)
	})
	t.Run("invalid zip file", func(t *testing.T) {
		dest := helpers.TempDir(t)
		defer helpers.DeleteDir(t, dest)

		_, err := Unzip(filepath.Join("tests", "archive4.zip"), dest)

		assert.Error(t, err)
	})
}
