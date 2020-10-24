package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests"
)

func Test_ReadConfig(t *testing.T) {
	tests := []struct {
		setup         func() (string, tests.TearDown)
		want          Config
		errorExpected bool
	}{
		{
			setup: func() (string, tests.TearDown) {
				return ".", tests.NoopTeardown()
			},
			want:          Config{},
			errorExpected: true,
		},
		{
			setup: func() (string, tests.TearDown) {
				return "", tests.NoopTeardown()
			},
			want:          Config{},
			errorExpected: true,
		},
		{
			setup: func() (string, tests.TearDown) {
				file := tests.TempFile(t, "", []byte("hello world"))
				return file, tests.DeleteFile(t, file)
			},
			want:          Config{},
			errorExpected: true,
		},
		{
			setup: func() (string, tests.TearDown) {
				file := tests.TempFile(t, "", []byte{})
				return file, tests.DeleteFile(t, file)
			},
			want:          Config{},
			errorExpected: false,
		},
		{
			setup: func() (string, tests.TearDown) {
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
				file := tests.TempFile(t, "", content)
				return file, tests.DeleteFile(t, file)
			},
			want: Config{
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
			},
			errorExpected: false,
		},
	}

	for _, tt := range tests {
		file, teardown := tt.setup()

		cfg, err := ReadConfig(file)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, cfg)
		}

		teardown()
	}
}

func Test_CreateDefaultConfig(t *testing.T) {
	dir := tests.TempDir(t)
	defer tests.DeleteDir(t, dir)()

	tests := []struct {
		setup         func() (string, tests.TearDown)
		errorExpected bool
	}{
		{
			setup: func() (string, tests.TearDown) {
				file := filepath.Join(dir, "file1")
				return file, tests.DeleteFile(t, file)
			},
			errorExpected: false,
		},
		{
			setup: func() (string, tests.TearDown) {
				file := tests.TempFile(t, dir, []byte{})
				assert.NoError(t, os.Chmod(file, os.FileMode(0400)))
				return file, tests.DeleteFile(t, file)
			},
			errorExpected: true,
		},
	}

	for _, tt := range tests {
		path, teardown := tt.setup()
		err := CreateDefaultConfig(path)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		teardown()
	}
}
