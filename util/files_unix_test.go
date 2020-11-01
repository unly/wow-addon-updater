// +build !windows

package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/unly/wow-addon-updater/util/tests/helpers"
)

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
			hidden, err := IsHiddenFile(actual)
			assert.NoError(t, err)
			assert.True(t, hidden)
		}

		tt.teardown()
	}
}

func Test_IsHiddenFilePath(t *testing.T) {
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
		actual := IsHiddenFilePath(tt.path)
		assert.Equal(t, tt.want, actual)
	}
}

func Test_WriteToHiddenFile_Unix(t *testing.T) {
	err := WriteToHiddenFile("dir/test.file", []byte("hello world"), os.FileMode(0666))
	assert.Error(t, err, "WriteToHiddenFile() returned no error")
}
