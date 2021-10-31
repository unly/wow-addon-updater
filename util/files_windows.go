//go:build windows

package util

import (
	"os"
	"syscall"
)

// HideFile hides the file at the given path.
// from https://stackoverflow.com/questions/54139606/how-to-create-a-hidden-file-in-windows-mac-linux
func HideFile(path string) (string, error) {
	return path, setFileAttribute(path, syscall.FILE_ATTRIBUTE_HIDDEN)
}

// WriteToHiddenFile writes the given data to a hidden file.
func WriteToHiddenFile(path string, data []byte, perm os.FileMode) error {
	if hidden, err := IsHiddenFile(path); err == nil && hidden {
		if err := setFileAttribute(path, syscall.FILE_ATTRIBUTE_NORMAL); err != nil {
			return err
		}
	}

	err := os.WriteFile(path, data, perm)
	if err != nil {
		return err
	}

	_, err = HideFile(path)

	return err
}

// IsHiddenFile returns whether the given file path is a hidden file or not.
func IsHiddenFile(path string) (bool, error) {
	if !FileExists(path) {
		return false, nil
	}

	filenameW, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return false, err
	}

	attrs, err := syscall.GetFileAttributes(filenameW)

	return attrs&syscall.FILE_ATTRIBUTE_HIDDEN != 0, err
}

// IsHiddenFilePath returns whether the given path could be a hidden file in the os.
// Always true for Windows.
func IsHiddenFilePath(_ string) bool {
	return true
}

func setFileAttribute(path string, attribute uint32) error {
	if !FileExists(path) {
		return os.ErrNotExist
	}

	filenameW, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	mode := info.Mode()
	err = syscall.SetFileAttributes(filenameW, attribute)
	if err2 := os.Chmod(path, mode); err != nil {
		err = err2
	}

	return err
}
