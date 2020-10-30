package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func Test_ReadConfig(t *testing.T) {
	type readConfigTest struct {
		file          string
		want          Config
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *readConfigTest{
		func() *readConfigTest {
			return &readConfigTest{
				file:          ".",
				want:          Config{},
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *readConfigTest {
			return &readConfigTest{
				file:          "",
				want:          Config{},
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *readConfigTest {
			file := helpers.TempFile(t, "", []byte("hello world"))
			return &readConfigTest{
				file:          file,
				want:          Config{},
				errorExpected: true,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
		func() *readConfigTest {
			file := helpers.TempFile(t, "", []byte{})
			return &readConfigTest{
				file:          file,
				want:          Config{},
				errorExpected: false,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
		func() *readConfigTest {
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
			return &readConfigTest{
				file: file,
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
				teardown:      helpers.DeleteFile(t, file),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		cfg, err := ReadConfig(tt.file)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, cfg)
		}

		tt.teardown()
	}
}

func Test_CreateDefaultConfig(t *testing.T) {
	type createDefaultConfigTest struct {
		file          string
		errorExpected bool
		teardown      helpers.TearDown
	}

	dir := helpers.TempDir(t)
	defer helpers.DeleteDir(t, dir)()

	tests := []func() *createDefaultConfigTest{
		func() *createDefaultConfigTest {
			file := filepath.Join(dir, "file1")
			return &createDefaultConfigTest{
				file:          file,
				errorExpected: false,
				teardown:      helpers.DeleteDir(t, file),
			}
		},
		func() *createDefaultConfigTest {
			file := helpers.TempFile(t, dir, []byte{})
			assert.NoError(t, os.Chmod(file, os.FileMode(0400)))
			return &createDefaultConfigTest{
				file:          file,
				errorExpected: true,
				teardown:      helpers.DeleteDir(t, file),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()
		err := CreateDefaultConfig(tt.file)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
		}

		tt.teardown()
	}
}
