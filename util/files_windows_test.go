// +build windows

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func Test_HideFile(t *testing.T) {
	type hideFileTest struct {
		path          string
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *hideFileTest{
		func() *hideFileTest {
			dir := helpers.TempDir(t)
			f := filepath.Join(dir, "test1")
			err := ioutil.WriteFile(f, []byte{}, os.FileMode(0666))
			assert.NoError(t, err)

			return &hideFileTest{
				path:          f,
				errorExpected: false,
				teardown:      helpers.DeleteDir(t, dir),
			}
		},
		func() *hideFileTest {
			dir := helpers.TempDir(t)
			f := filepath.Join(dir, ".test2")
			err := ioutil.WriteFile(f, []byte{}, os.FileMode(0666))
			assert.NoError(t, err)
			err = setFileAttribute(f, syscall.FILE_ATTRIBUTE_HIDDEN)
			assert.NoError(t, err)

			return &hideFileTest{
				path:          f,
				errorExpected: false,
				teardown:      helpers.DeleteDir(t, dir),
			}
		},
		func() *hideFileTest {
			return &hideFileTest{
				path:          "",
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *hideFileTest {
			return &hideFileTest{
				path:          ".",
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *hideFileTest {
			return &hideFileTest{
				path:          "fake.file",
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		actual, err := HideFile(tt.path)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.path, actual)
			hidden, err := IsHiddenFile(tt.path)
			assert.NoError(t, err)
			assert.True(t, hidden)
		}

		tt.teardown()
	}
}

func Test_IsHiddenFilePath(t *testing.T) {
	tests := []string{
		"",
		".",
		"file",
		".file",
	}

	for _, tt := range tests {
		actual := IsHiddenFilePath(tt)
		assert.True(t, actual)
	}
}
