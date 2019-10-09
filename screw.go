package screw

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

var DEBUG = os.Getenv("DEBUG_SCREW") == "1"

func Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

func Create(name string) (*os.File, error) {
	return os.Create(name)
}

func Open(name string) (*os.File, error) {
	return OpenFile(name, os.O_RDONLY, 0)
}

func OpenFile(path string, flag int, perm os.FileMode) (*os.File, error) {
	if !IsActualCasing(path) {
		return nil, os.ErrNotExist
	}

	return os.OpenFile(path, flag, perm)
}

func IsActualCasing(path string) (bool, error) {
	debugf("testing (%s)", path)

	var data windows.Win32finddata
	utf16Str, err := windows.UTF16FromString(path)
	if err != nil {
		debugf("utf16 error: +%v", err)
		return false, err
	}

	_, err = windows.FindFirstFile(&utf16Str[0], &data)
	if err != nil {
		debugf("file not found: %+v", err)
		return false, err
	}
	name := windows.UTF16ToString(data.FileName[:259])

	reqpath := filepath.Base(path)
	actpath := name

	if reqpath != actpath {
		debugf("requested (%s) != actual (%s)", reqpath, actpath)
		return false, err
	}

	dir := filepath.Dir(path)
	// true for `/`, for `C:\`, etc.
	if dir == filepath.Dir(dir) {
		return true, nil
	} else {
		return IsActualCasing(dir)
	}
}

func debugf(f string, arg ...interface{}) {
	if DEBUG {
		fmt.Printf("[screw] %s\n", fmt.Sprintf(f, arg...))
	}
}
