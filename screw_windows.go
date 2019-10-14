// +build windows

package screw

import (
	"os"

	"golang.org/x/sys/windows"
)

func doRename(oldpath, newpath string) error {
	var lastErr error
	sleeper := newSleeper()

	for {
		err := os.Rename(oldpath, newpath)
		lastErr = err
		if err == nil {
			break
		}

		if !os.IsNotExist(err) && sleeper.Sleep(err) {
			continue
		}
		break
	}
	return lastErr
}

func TrueBaseName(name string) string {
	var data windows.Win32finddata
	utf16Str, err := windows.UTF16FromString(name)
	if err != nil {
		return ""
	}

	_, err = windows.FindFirstFile(&utf16Str[0], &data)
	if err != nil {
		return ""
	}

	return windows.UTF16ToString(data.FileName[:windows.MAX_PATH-1])
}
