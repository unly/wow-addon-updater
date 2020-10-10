// +build windows

package util

import "syscall"

// from https://stackoverflow.com/questions/54139606/how-to-create-a-hidden-file-in-windows-mac-linux
func HideFile(filename string) error {
	filenameW, err := syscall.UTF16PtrFromString(filename)
	if err != nil {
		return err
	}

	return syscall.SetFileAttributes(filenameW, syscall.FILE_ATTRIBUTE_HIDDEN)
}
