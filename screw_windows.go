// +build windows

package screw

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

func Rename(oldpath, newpath string) error {
	if strings.ToLower(oldpath) == strings.ToLower(newpath) {
		tmppath := oldpath + fmt.Sprintf("_rename_%d", os.Getpid())
		err := doRename(oldpath, tmppath)
		if err != nil {
			return err
		}

		err = doRename(tmppath, newpath)
		if err != nil {
			// wouldn't really know what to do with a second error here
			_ = doRename(tmppath, oldpath)
			return err
		}

		return nil
	}

	return os.Rename(oldpath, newpath)
}

func doRename(oldpath, newpath string) error {
	sleepIntervals := []int{0, 250, 1000, 5 * 1000, 10 * 1000}
	var lastErr error

	for {
		err := os.Rename(oldpath, newpath)
		lastErr = err
		if err == nil {
			break
		}

		if len(sleepIntervals) > 0 && os.IsPermission(err) {
			s := sleepIntervals[0]
			sleepIntervals = sleepIntervals[1:]

			sleepTime := time.Duration(s) * time.Millisecond
			debugf("sleeping %v because: %+v", sleepTime, err)
			time.Sleep(sleepTime)
			continue
		}
		break
	}
	return lastErr
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	actualCasing, err := IsActualCasing(name)
	if err != nil {
		if os.IsNotExist(err) && (flag&os.O_CREATE) > 0 {
			// that's ok
		} else {
			return nil, err
		}
	} else {
		if !actualCasing {
			return nil, os.ErrNotExist
		}
	}

	return os.OpenFile(name, flag, perm)
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
