package helpers

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TearDown func()

func DeleteFile(t *testing.T, path string) TearDown {
	return func() {
		t.Helper()
		assert.NoErrorf(t, os.Remove(path), "failed to delete file: %s", path)
	}
}

func DeleteDir(t *testing.T, path string) TearDown {
	return func() {
		t.Helper()
		assert.NoErrorf(t, os.RemoveAll(path), "failed to delete directory: %s", path)
	}
}

func NoopTeardown() TearDown {
	return func() {}
}

func TempDir(t *testing.T) string {
	t.Helper()
	dir, err := ioutil.TempDir("", "*")
	assert.NoError(t, err, "failed to create a temp directory")
	return dir
}

func TempFile(t *testing.T, dir string, content []byte) string {
	t.Helper()
	f, err := ioutil.TempFile(dir, "*")
	assert.NoError(t, err, "failed to create temporary file")
	defer f.Close()
	_, err = f.Write(content)
	assert.NoErrorf(t, err, "failed to write content to: %s", f.Name())
	return f.Name()
}
