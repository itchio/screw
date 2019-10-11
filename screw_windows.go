// +build windows

package screw

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows"
)

func Mkdir(name string, perm os.FileMode) error {
	wrap := mkwrap("screw.Mkdir", name)

	err := os.Mkdir(name, perm)
	if err != nil {
		return err
	}

	actualCase, err := IsActualCasing(name)
	if err != nil {
		return wrap(err)
	}
	if !actualCase {
		return wrap(ErrCaseConflict)
	}
	return nil
}

func MkdirAll(name string, perm os.FileMode) error {
	wrap := mkwrap("screw.MkdirAll", name)

	return wrap(errors.New("stub!"))
}

func Rename(oldpath, newpath string) error {
	wrap := mkwrap("screw.Rename", oldpath)

	err := doRename(oldpath, newpath)
	if err != nil {
		return wrap(err)
	}

	// casing changed?
	if strings.ToLower(oldpath) == strings.ToLower(newpath) {
		// was it changed properly?
		isActual, err := IsActualCasing(newpath)
		if err != nil {
			return wrap(err)
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
				return wrap(err)
			}

			err = doRename(tmppath, newpath)
			if err != nil {
				return wrap(err)
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
	wrap := mkwrap("screw.OpenFile", name)

	actualCasing, err := IsActualCasing(name)
	if err != nil {
		if os.IsNotExist(err) && (flag&os.O_CREATE) > 0 {
			// that's ok
		} else {
			return nil, wrap(err)
		}
	} else {
		if !actualCasing {
			if (flag & os.O_CREATE) > 0 {
				return nil, wrap(ErrCaseConflict)
			} else {
				return nil, wrap(os.ErrNotExist)
			}
		}
	}

	f, err := os.OpenFile(name, flag, perm)
	if err != nil {
		return nil, wrap(err)
	}
	return f, nil
}

func Stat(name string) (os.FileInfo, error) {
	wrap := mkwrap("screw.Stat", name)

	actualCasing, err := IsActualCasing(name)
	if err != nil {
		return nil, err
	}

	if !actualCasing {
		return nil, wrap(os.ErrNotExist)
	}

	stats, err := os.Stat(name)
	if err != nil {
		return nil, wrap(err)
	}
	return stats, nil
}

func Lstat(name string) (os.FileInfo, error) {
	wrap := mkwrap("screw.Lstat", name)

	actualCasing, err := IsActualCasing(name)
	if err != nil {
		return nil, err
	}

	if !actualCasing {
		return nil, wrap(os.ErrNotExist)
	}

	stats, err := os.Lstat(name)
	if err != nil {
		return nil, wrap(err)
	}
	return stats, nil
}

func RemoveAll(name string) error {
	wrap := mkwrap("screw.RemoveAll", name)

	isActual, err := IsActualCasing(name)
	if err != nil {
		if os.IsNotExist(err) {
			// neither "apricot", "APRICOT", etc. exist
			return nil
		}
		return wrap(err)
	}

	if !isActual {
		// asked to remove "apricot" but "APRICOT" exists, consider already removed
		return nil
	}

	// accepting to try and remove "apricot" and all its children
	err = os.RemoveAll(name)
	if err != nil {
		return wrap(err)
	}
	return nil
}

func Remove(name string) error {
	wrap := mkwrap("screw.Remove", name)

	isActual, err := IsActualCasing(name)
	if err != nil {
		if os.IsNotExist(err) {
			// neither "apricot", "APRICOT", etc. exist
			return wrap(os.ErrNotExist)
		}
		return (err)
	}

	if !isActual {
		// asked to remove "apricot" but "APRICOT" (or another case variant)
		// exists, so, can't remove "apricot" because it doesn't exist
		return wrap(os.ErrNotExist)
	}

	// accepting to try and remove "apricot"
	err = os.Remove(name)
	if err != nil {
		return wrap(err)
	}
	return nil
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

func mkwrap(op string, path string) func(err error) error {
	return func(err error) error {
		return wrap(err, op, path)
	}
}

func wrap(err error, op string, path string) error {
	return &os.PathError{
		Op:   op,
		Path: path,
		Err:  err,
	}
}
