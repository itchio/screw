// +build windows

package screw

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/sys/windows"
)

func Mkdir(name string, perm os.FileMode) error {
	wrap := mkwrap("screw.Mkdir", name)

	err := os.Mkdir(name, perm)
	if err != nil {
		mkdirErr := err
		if os.IsExist(mkdirErr) {
			actualCase, err := IsActualCasing(name)
			if err != nil {
				return wrap(err)
			}
			if !actualCase {
				return wrap(ErrCaseConflict)
			}
		}

		return mkdirErr
	}
	return nil
}

func MkdirAll(path string, perm os.FileMode) error {
	wrap := mkwrap("screw.MkdirAll", path)

	// modelled after `go/src/os/path.go`

	// Fast path: if we can tell whether path is a directory or file, stop with success or error.
	dir, err := Stat(path)
	if err == nil {
		if dir.IsDir() {
			return nil
		}
		return wrap(syscall.ENOTDIR)
	}

	// Slow path: make sure parent exists and then call Mkdir for path.
	i := len(path)
	for i > 0 && os.IsPathSeparator(path[i-1]) { // Skip trailing path separator.
		i--
	}

	j := i
	for j > 0 && !os.IsPathSeparator(path[j-1]) { // Scan backward over element.
		j--
	}

	if j > 1 {
		// Create parent.
		err = MkdirAll(fixRootDirectory(path[:j-1]), perm)
		if err != nil {
			return wrap(err)
		}
	}

	// Parent now exists; invoke Mkdir and use its result.
	err = Mkdir(path, perm)
	if err != nil {
		// Handle arguments like "foo/." by
		// double-checking that directory doesn't exist.
		dir, err1 := Lstat(path)
		if err1 == nil && dir.IsDir() {
			return nil
		}
		return wrap(err)
	}
	return nil
}

// fixRootDirectory fixes a reference to a drive's root directory to
// have the required trailing slash.
func fixRootDirectory(p string) string {
	if len(p) == len(`\\?\c:`) {
		if os.IsPathSeparator(p[0]) && os.IsPathSeparator(p[1]) && p[2] == '?' && os.IsPathSeparator(p[3]) && p[5] == ':' {
			return p + `\`
		}
	}
	return p
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

func IsActualCasing(name string) (bool, error) {
	absolutePath, err := filepath.Abs(name)
	if err != nil {
		return false, err
	}
	return isActualCasing(absolutePath)
}

func isActualCasing(absolutePath string) (bool, error) {
	actualBaseName, err := getActualBaseName(absolutePath)
	if err != nil {
		return false, err
	}

	if filepath.Base(absolutePath) != actualBaseName {
		return false, err
	}

	absoluteParentPath := filepath.Dir(absolutePath)

	// true for `C:\`, etc.
	if absoluteParentPath == filepath.Dir(absoluteParentPath) {
		return true, nil
	} else {
		return isActualCasing(absoluteParentPath)
	}
}

func GetActualCasing(name string) (string, error) {
	absolutePath, err := filepath.Abs(name)
	if err != nil {
		return "", err
	}
	return getActualCasing(absolutePath)
}

func getActualCasing(absolutePath string) (string, error) {
	absoluteParentPath := filepath.Dir(absolutePath)
	actualBaseName, err := getActualBaseName(absolutePath)
	if err != nil {
		return "", err
	}

	if filepath.Dir(absoluteParentPath) == absoluteParentPath {
		// At this point, `absoluteParentPath` is `C:\`, or a casing
		// variant thereof. The canonical version of drive letters is
		// uppercase, so let's force it now just in case
		drive := strings.ToUpper(absoluteParentPath)
		return filepath.Join(drive, actualBaseName), nil
	} else {
		actualParent, err := getActualCasing(absoluteParentPath)
		if err != nil {
			return "", err
		}
		return filepath.Join(actualParent, actualBaseName), nil
	}
}

func getActualBaseName(path string) (string, error) {
	var data windows.Win32finddata
	utf16Str, err := windows.UTF16FromString(path)
	if err != nil {
		return "", err
	}

	_, err = windows.FindFirstFile(&utf16Str[0], &data)
	if err != nil {
		return "", err
	}

	return windows.UTF16ToString(data.FileName[:windows.MAX_PATH-1]), nil
}
