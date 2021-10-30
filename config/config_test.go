package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func TestReadConfig(t *testing.T) {
	t.Run("current dir", func(t *testing.T) {
		_, err := ReadConfig(".")

		assert.Error(t, err)
	})
	t.Run("empty string", func(t *testing.T) {
		_, err := ReadConfig("")

		assert.Error(t, err)
	})
	t.Run("invalid file content", func(t *testing.T) {
		file := helpers.TempFile(t, "", []byte("hello world"))
		defer helpers.DeleteFile(t, file)

		_, err := ReadConfig(file)

		assert.Error(t, err)
	})
	t.Run("empty file", func(t *testing.T) {
		file := helpers.TempFile(t, "", []byte{})
		defer helpers.DeleteFile(t, file)

		cfg, err := ReadConfig(file)

		assert.NoError(t, err)
		assert.Equal(t, Config{}, cfg)
	})
	t.Run("sample config", func(t *testing.T) {
		content := []byte(`
classic:
  path: path/to/classic
  addons:
    - addon1
    - addon2
retail:
  path: path/to/retail
  addons:
    - addon3
    - addon4`)
		file := helpers.TempFile(t, "", content)
		defer helpers.DeleteFile(t, file)

		cfg, err := ReadConfig(file)

		assert.NoError(t, err)
		want := Config{
			Classic: WowConfig{
				Path: "path/to/classic",
				AddOns: []string{
					"addon1",
					"addon2",
				},
			},
			Retail: WowConfig{
				Path: "path/to/retail",
				AddOns: []string{
					"addon3",
					"addon4",
				},
			},
		}
		assert.Equal(t, want, cfg)
	})
}

func TestCreateDefaultConfig(t *testing.T) {
	dir := helpers.TempDir(t)
	defer helpers.DeleteDir(t, dir)()

	t.Run("create default config", func(t *testing.T) {
		file := filepath.Join(dir, "file1")

		err := CreateDefaultConfig(file)

		assert.NoError(t, err)
		helpers.DeleteDir(t, file)
	})
	t.Run("existing read only file", func(t *testing.T) {
		file := helpers.TempFile(t, dir, []byte{})
		err := os.Chmod(file, os.FileMode(0400))
		if err != nil {
			assert.FailNow(t, "failed to chmod 04400")
		}

		err = CreateDefaultConfig(file)

		assert.Error(t, err)
		helpers.DeleteDir(t, file)
	})
}
