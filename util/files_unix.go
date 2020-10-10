// +build !windows

package util

import (
	"os"
	"path/filepath"
	"strings"
)

func HideFile(path string) error {
	if filename := filepath.Base(path); !strings.HasPrefix(filename, ".") {
		newPath := filepath.Join(filepath.Dir(path), "."+filename)

		return os.Rename(path, newPath)
	}

	return nil
}
