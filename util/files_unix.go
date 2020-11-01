// +build !windows

package util

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// HideFile hides the file at the given path.
// Might return another path with a leading . in the filename.
func HideFile(path string) (string, error) {
	if !FileExists(path) {
		return "", os.ErrNotExist
	}

	if filename := filepath.Base(path); !strings.HasPrefix(filename, ".") {
		newPath := filepath.Join(filepath.Dir(path), "."+filename)

		return newPath, os.Rename(path, newPath)
	}

	return path, nil
}

// WriteToHiddenFile writes the given data to a hidden file.
func WriteToHiddenFile(path string, data []byte, perm os.FileMode) error {
	if !IsHiddenFilePath(path) {
		return fmt.Errorf("the path %s is not valid for a hidden file", path)
	}

	return ioutil.WriteFile(path, data, perm)
}

// IsHiddenFile returns whether the given file path is a hidden file or not.
func IsHiddenFile(path string) (bool, error) {
	if !FileExists(path) {
		return false, nil
	}

	return IsHiddenFilePath(path), nil
}

// IsHiddenFilePath returns whether the given path could be a hidden file in the os.
func IsHiddenFilePath(path string) bool {
	filename := filepath.Base(path)
	return len(filename) > 1 && strings.HasPrefix(filename, ".")
}
