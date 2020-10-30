// +build !windows

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

func isHidden(t *testing.T, path string) {
	t.Helper()
	assert.True(t, strings.HasPrefix(filepath.Base(path), "."))
}

func Test_HideFile(t *testing.T) {
	type hideFileTest struct {
		path          string
		returnedPath  string
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
				returnedPath:  filepath.Join(dir, ".test1"),
				errorExpected: false,
				teardown:      helpers.DeleteDir(t, dir),
			}
		},
		func() *hideFileTest {
			dir := helpers.TempDir(t)
			f := filepath.Join(dir, ".test2")
			err := ioutil.WriteFile(f, []byte{}, os.FileMode(0666))
			assert.NoError(t, err)

			return &hideFileTest{
				path:          f,
				returnedPath:  filepath.Join(dir, ".test2"),
				errorExpected: false,
				teardown:      helpers.DeleteDir(t, dir),
			}
		},
		func() *hideFileTest {
			return &hideFileTest{
				path:          "",
				returnedPath:  "",
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *hideFileTest {
			return &hideFileTest{
				path:          ".",
				returnedPath:  "",
				errorExpected: true,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *hideFileTest {
			return &hideFileTest{
				path:          "fake.file",
				returnedPath:  "",
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
			assert.Equal(t, tt.returnedPath, actual)
			isHidden(t, actual)
		}

		tt.teardown()
	}
}
