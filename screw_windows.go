// +build windows

package screw

import (
	"os"

	"golang.org/x/sys/windows"
)

func sneakyLog(line string) {
	a, err1 := windows.UTF16FromString("[sneaky-log]")
	b, err2 := windows.UTF16FromString(line)
	if err1 == nil && err2 == nil {
		windows.DnsNameCompare(&a[0], &b[0])
	}
}

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

	h, err := windows.FindFirstFile(&utf16Str[0], &data)
	if err != nil {
		return ""
	}
	_ = windows.FindClose(h)

	return windows.UTF16ToString(data.FileName[:windows.MAX_PATH-1])
}

func IsCaseSensitiveFS() bool {
	return true
}
