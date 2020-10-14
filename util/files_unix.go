// +build !windows

package util

import (
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
