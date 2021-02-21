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
	assert.NoError(t, err)
}

func setFileWriteable(t *testing.T, path string) {
	t.Helper()
	err := os.Chmod(path, os.FileMode(0600))
	assert.NoError(t, err)
}

func Test_FileExists(t *testing.T) {
	type fileExistsTest struct {
		path     string
		want     bool
		teardown helpers.TearDown
	}

	tests := []func() *fileExistsTest{
		func() *fileExistsTest {
			file := helpers.TempFile(t, "", []byte{})

			return &fileExistsTest{
				path:     file,
				want:     true,
				teardown: helpers.DeleteFile(t, file),
			}
		},
		func() *fileExistsTest {
			dir := helpers.TempDir(t)

			return &fileExistsTest{
				path:     dir,
				want:     false,
				teardown: helpers.DeleteDir(t, dir),
			}
		},
		func() *fileExistsTest {
			return &fileExistsTest{
				path:     "",
				want:     false,
				teardown: helpers.NoopTeardown(),
			}
		},
		func() *fileExistsTest {
			return &fileExistsTest{
				path:     ".",
				want:     false,
				teardown: helpers.NoopTeardown(),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		actual := FileExists(tt.path)

		assert.Equal(t, tt.want, actual, "FileExists() returned %+v, want %+v", actual, tt.want)

		tt.teardown()
	}
}

func Test_IsHiddenFile(t *testing.T) {
	type isHiddenFileTest struct {
		path          string
		want          bool
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *isHiddenFileTest{
		func() *isHiddenFileTest {
			return &isHiddenFileTest{
				path:          "",
				want:          false,
				errorExpected: false,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *isHiddenFileTest {
			return &isHiddenFileTest{
				path:          ".",
				want:          false,
				errorExpected: false,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *isHiddenFileTest {
			return &isHiddenFileTest{
				path:          "fake file",
				want:          false,
				errorExpected: false,
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *isHiddenFileTest {
			file := helpers.TempFile(t, "", []byte{})

			return &isHiddenFileTest{
				path:          file,
				want:          false,
				errorExpected: false,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
		func() *isHiddenFileTest {
			file, err := HideFile(helpers.TempFile(t, "", []byte{}))
			assert.NoError(t, err)

			return &isHiddenFileTest{
				path:          file,
				want:          true,
				errorExpected: false,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		hidden, err := IsHiddenFile(tt.path)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.want, hidden)
		}

		tt.teardown()
	}
}

func Test_WriteToHiddenFile(t *testing.T) {
	type writeToHiddenFileTest struct {
		path          string
		data          []byte
		perm          os.FileMode
		errorExpected bool
		teardown      helpers.TearDown
	}

	tests := []func() *writeToHiddenFileTest{
		func() *writeToHiddenFileTest {
			dir := helpers.TempDir(t)
			file := filepath.Join(dir, ".test")

			return &writeToHiddenFileTest{
				path:          file,
				data:          []byte("hello world"),
				perm:          os.FileMode(0666),
				errorExpected: false,
				teardown:      helpers.DeleteDir(t, dir),
			}
		},
		func() *writeToHiddenFileTest {
			file, err := HideFile(helpers.TempFile(t, "", []byte{}))
			assert.NoError(t, err)

			return &writeToHiddenFileTest{
				path:          file,
				data:          []byte("hello world"),
				perm:          os.FileMode(0666),
				errorExpected: false,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
		func() *writeToHiddenFileTest {
			file, err := HideFile(helpers.TempFile(t, "", []byte{}))
			assert.NoError(t, err)

			return &writeToHiddenFileTest{
				path:          file,
				data:          []byte{},
				perm:          os.FileMode(0666),
				errorExpected: false,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
		func() *writeToHiddenFileTest {
			file, err := HideFile(helpers.TempFile(t, "", []byte{}))
			assert.NoError(t, err)
			setFileReadOnly(t, file)

			return &writeToHiddenFileTest{
				path:          file,
				data:          []byte("i should not be there"),
				perm:          os.FileMode(0666),
				errorExpected: true,
				teardown: func() {
					setFileWriteable(t, file)
					helpers.DeleteFile(t, file)()
				},
			}
		},
		func() *writeToHiddenFileTest {
			file, err := HideFile(helpers.TempFile(t, "", []byte("old content")))
			assert.NoError(t, err)

			return &writeToHiddenFileTest{
				path:          file,
				data:          []byte("new content"),
				perm:          os.FileMode(0666),
				errorExpected: false,
				teardown:      helpers.DeleteFile(t, file),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		err := WriteToHiddenFile(tt.path, tt.data, tt.perm)

		if tt.errorExpected {
			assert.Error(t, err, "WriteToHiddenFile() returned no error")
		} else {
			assert.NoError(t, err, "WriteToHiddenFile() returned error %+v", err)
			hidden, err := IsHiddenFile(tt.path)
			assert.NoError(t, err)
			assert.True(t, hidden)
			actualBody, err := os.ReadFile(tt.path)
			assert.NoError(t, err, "failed to read in temporary file %s", tt.path)
			assert.Equal(t, tt.data, actualBody, "WriteToHiddenFile() file content wrong")
		}

		_ = os.Remove(tt.path)
	}
}

func Test_Unzip(t *testing.T) {
	type unzipTest struct {
		src           string
		dest          string
		errorExpected bool
		files         []string
		teardown      helpers.TearDown
	}

	tests := []func() *unzipTest{
		func() *unzipTest {
			return &unzipTest{
				src:           "not-existing",
				dest:          "",
				errorExpected: true,
				files:         []string{},
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *unzipTest {
			return &unzipTest{
				src:           ".",
				dest:          "",
				errorExpected: true,
				files:         []string{},
				teardown:      helpers.NoopTeardown(),
			}
		},
		func() *unzipTest {
			dest := helpers.TempDir(t)

			return &unzipTest{
				src:           filepath.Join("tests", "archive1.zip"),
				dest:          dest,
				errorExpected: false,
				files: []string{
					filepath.Join(dest, "a.txt"),
					filepath.Join(dest, "b.txt"),
					filepath.Join(dest, "c.txt"),
				},
				teardown: helpers.DeleteDir(t, dest),
			}
		},
		func() *unzipTest {
			dest := helpers.TempDir(t)

			return &unzipTest{
				src:           filepath.Join("tests", "archive2.zip"),
				dest:          dest,
				errorExpected: false,
				files: []string{
					filepath.Join(dest, "a.txt"),
				},
				teardown: helpers.DeleteDir(t, dest),
			}
		},
		func() *unzipTest {
			dest := helpers.TempDir(t)

			return &unzipTest{
				src:           filepath.Join("tests", "archive3.zip"),
				dest:          dest,
				errorExpected: false,
				files:         nil,
				teardown:      helpers.DeleteDir(t, dest),
			}
		},
		func() *unzipTest {
			return &unzipTest{
				src:           filepath.Join("tests", "archive4.zip"),
				dest:          "",
				errorExpected: true,
				files:         []string{},
				teardown:      helpers.NoopTeardown(),
			}
		},
	}

	for _, fn := range tests {
		tt := fn()

		actual, err := Unzip(tt.src, tt.dest)

		if tt.errorExpected {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.files, actual)
		}

		tt.teardown()
	}
}
