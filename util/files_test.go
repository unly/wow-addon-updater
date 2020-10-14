package util

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func setFileReadOnly(t *testing.T, path string) {
	err := os.Chmod(path, os.FileMode(0400))
	assert.NoError(t, err)
}

func setFileWriteable(t *testing.T, path string) {
	err := os.Chmod(path, os.FileMode(0600))
	assert.NoError(t, err)
}

func Test_FileExists(t *testing.T) {
	tests := []struct {
		setup    func() string
		teardown func(path string)
		want     bool
	}{
		{
			setup: func() string {
				f, err := ioutil.TempFile("", "test.txt")
				assert.NoError(t, err)
				return f.Name()
			},
			teardown: func(path string) {
				os.Remove(path)
			},
			want: true,
		},
		{
			setup: func() string {
				dir, err := ioutil.TempDir("", "test-dir")
				assert.NoError(t, err)
				return dir
			},
			teardown: func(path string) {
				os.Remove(path)
			},
			want: false,
		},
		{
			setup: func() string {
				return ""
			},
			teardown: func(path string) {
			},
			want: false,
		},
		{
			setup: func() string {
				return "."
			},
			teardown: func(path string) {
			},
			want: false,
		},
	}

	for _, tt := range tests {
		path := tt.setup()

		actual := FileExists(path)

		assert.Equal(t, tt.want, actual, "FileExists() returned %+v, want %+v", actual, tt.want)

		tt.teardown(path)
	}
}

func Test_WriteToHiddenFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "hidden-tests")
	assert.NoError(t, err, "failed to initialize the temporary test directory %v", err)
	defer os.RemoveAll(dir)

	tests := []struct {
		path        func() string
		data        []byte
		perm        os.FileMode
		expectError bool
		teardown    func(path string)
	}{
		{
			path: func() string {
				return filepath.Join(dir, "test1.txt")
			},
			data:        []byte("hello world"),
			perm:        os.FileMode(0666),
			expectError: false,
			teardown: func(path string) {
				os.Remove(path)
			},
		},
		{
			path: func() string {
				return filepath.Join(dir, "test2.txt")
			},
			data:        []byte{},
			perm:        os.FileMode(0666),
			expectError: false,
			teardown: func(path string) {
				os.Remove(path)
			},
		},
		{
			path: func() string {
				f, err := ioutil.TempFile(dir, "test3")
				assert.NoError(t, err)
				setFileReadOnly(t, f.Name())
				return f.Name()
			},
			data:        []byte("i should not be there"),
			perm:        os.FileMode(0666),
			expectError: true,
			teardown: func(path string) {
				setFileWriteable(t, path)
				os.Remove(path)
			},
		},
		{
			path: func() string {
				filename := filepath.Join(dir, "test4")
				err := ioutil.WriteFile(filename, []byte("old content"), os.FileMode(0666))
				assert.NoError(t, err)
				return filename
			},
			data:        []byte("new content"),
			perm:        os.FileMode(0666),
			expectError: false,
			teardown: func(path string) {
				os.Remove(path)
			},
		},
	}

	for _, tt := range tests {
		filePath := tt.path()
		path, actual := WriteToHiddenFile(filePath, tt.data, tt.perm)

		if tt.expectError {
			assert.Error(t, actual, "WriteToHiddenFile() returned no error")
			_, err := ioutil.ReadFile(path)
			assert.Error(t, err)
			tt.teardown(filePath)
		} else {
			assert.NoError(t, actual, "WriteToHiddenFile() returned error %+v", actual)
			isHidden(t, path)
			actualBody, err := ioutil.ReadFile(path)
			assert.NoError(t, err, "failed to read in temporary file %s", path)
			assert.Equal(t, tt.data, actualBody, "WriteToHiddenFile() file content wrong")
			tt.teardown(path)
		}
	}
}

func Test_Unzip(t *testing.T) {
	dir, err := ioutil.TempDir("", "zip-tests")
	assert.NoError(t, err, "failed to initialize the temporary test directory %v", err)
	defer os.RemoveAll(dir)

	tests := []struct {
		src           func() string
		dest          func() string
		expectedError bool
		files         func(dest string) []string
		teardown      func(src, dest string)
	}{
		{
			src: func() string {
				return "not-existing"
			},
			dest: func() string {
				return ""
			},
			expectedError: true,
			files: func(dest string) []string {
				return []string{}
			},
			teardown: func(src, dest string) {},
		},
		{
			src: func() string {
				return dir
			},
			dest: func() string {
				return ""
			},
			expectedError: true,
			files: func(dest string) []string {
				return []string{}
			},
			teardown: func(src, dest string) {},
		},
		{
			src: func() string {
				return filepath.Join("tests", "archive1.zip")
			},
			dest: func() string {
				d, err := ioutil.TempDir(dir, "archive")
				assert.NoError(t, err)
				return d
			},
			expectedError: false,
			files: func(dest string) []string {
				return []string{
					filepath.Join(dest, "a.txt"),
					filepath.Join(dest, "b.txt"),
					filepath.Join(dest, "c.txt"),
				}
			},
			teardown: func(src, dest string) {
				os.RemoveAll(dest)
			},
		},
		{
			src: func() string {
				return filepath.Join("tests", "archive2.zip")
			},
			dest: func() string {
				d, err := ioutil.TempDir(dir, "archive")
				assert.NoError(t, err)
				return d
			},
			expectedError: false,
			files: func(dest string) []string {
				return []string{
					filepath.Join(dest, "a.txt"),
				}
			},
			teardown: func(src, dest string) {
				os.RemoveAll(dest)
			},
		},
		{
			src: func() string {
				return filepath.Join("tests", "archive3.zip")
			},
			dest: func() string {
				d, err := ioutil.TempDir(dir, "archive")
				assert.NoError(t, err)
				return d
			},
			expectedError: false,
			files: func(dest string) []string {
				return nil
			},
			teardown: func(src, dest string) {
				os.RemoveAll(dest)
			},
		},
		{
			src: func() string {
				return filepath.Join("tests", "archive4.zip")
			},
			dest: func() string {
				return ""
			},
			expectedError: true,
			files: func(dest string) []string {
				return []string{}
			},
			teardown: func(src, dest string) {},
		},
	}

	for _, tt := range tests {
		src := tt.src()
		dest := tt.dest()

		actual, err := Unzip(src, dest)

		if tt.expectedError {
			assert.Error(t, err)
		} else {
			assert.NoError(t, err)
			assert.Equal(t, tt.files(dest), actual)
		}

		tt.teardown(src, dest)
	}
}
