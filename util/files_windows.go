// +build windows

package util

import (
	"os"
	"syscall"
)

// HideFile hides the file at the given path.
// Returns the same path and potentially an error.
// from https://stackoverflow.com/questions/54139606/how-to-create-a-hidden-file-in-windows-mac-linux
func HideFile(path string) (string, error) {
	if !FileExists(path) {
		return "", os.ErrNotExist
	}

	filenameW, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return "", err
	}

	return path, syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
}
