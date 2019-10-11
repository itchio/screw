// +build windows

package screw

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func Rename(oldpath, newpath string) error {
	err := doRename(oldpath, newpath)
	if err != nil {
		return err
	}

	// casing changed?
	if strings.ToLower(oldpath) == strings.ToLower(newpath) {
		// was it changed properly?
		isActual, err := IsActualCasing(newpath)
		if err != nil {
			return &os.PathError{
				Op:   "screw.Rename",
				Path: newpath,
				Err:  ErrWrongCasing,
			}
		}

		if isActual {
			// thankfully, yes!
			return nil
		}

		if !isActual {
			tmppath := oldpath + fmt.Sprintf("_rename_%d", os.Getpid())
			err := doRename(oldpath, tmppath)
			if err != nil {
				// here is an awkward place to return an error, but
				// if anyone has a better idea, I'm listening.. :(
				return err
			}

			err = doRename(tmppath, newpath)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return nil
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
			return nil, &os.PathError{
				Op:   "screw.OpenFile",
				Path: name,
				Err:  ErrWrongCasing,
			}
		}
	}

	return os.OpenFile(name, flag, perm)
}

func Stat(name string) (os.FileInfo, error) {
	actualCasing, err := IsActualCasing(name)
	if err != nil {
		return nil, err
	}

	if !actualCasing {
		return nil, &os.PathError{
			Op:   "screw.Stat",
			Path: name,
			Err:  ErrWrongCasing,
		}
	}

	return os.Stat(name)
}

func Lstat(name string) (os.FileInfo, error) {
	actualCasing, err := IsActualCasing(name)
	if err != nil {
		return nil, err
	}

	if !actualCasing {
		return nil, &os.PathError{
			Op:   "screw.Stat",
			Path: name,
			Err:  ErrWrongCasing,
		}
	}

	return os.Lstat(name)
}

func IsActualCasing(path string) (bool, error) {
	path, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}

	return isActualCasing(path)
}

func isActualCasing(path string) (bool, error) {
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
	name := windows.UTF16ToString(data.FileName[:windows.MAX_PATH-1])

	reqpath := filepath.Base(path)
	actpath := name

	if reqpath != actpath {
		debugf("requested (%s) != actual (%s)", reqpath, actpath)
		return false, err
	}

	dir := filepath.Dir(path)
	// true for `C:\`, etc.
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
