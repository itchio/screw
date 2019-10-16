package screw

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var DEBUG = os.Getenv("DEBUG_SCREW") == "1"

var (
	ErrCaseConflict = errors.New("a file with a different case already exists on disk")
)

// Returns true if `name` exists on disk but
// with a different case.
// Returns false in any other case.
func IsWrongCase(name string) bool {
	name, err := filepath.Abs(name)
	if err != nil {
		return false
	}
	trueBase := TrueBaseName(name)
	if trueBase != "" && trueBase != filepath.Base(name) {
		return true
	}
	return false
}

func Create(name string) (*os.File, error) {
	debugf("screw.Create(%s)", name)
	return OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
}

func Open(name string) (*os.File, error) {
	debugf("screw.Open(%s)", name)
	return OpenFile(name, os.O_RDONLY, 0)
}

func Symlink(oldname string, newname string) error {
	debugf("screw.Symlink(%s, %s)", oldname, newname)
	return os.Symlink(oldname, newname)
}

func Truncate(name string, size int64) error {
	debugf("screw.Truncate(%s, %d)", name, size)
	wrap := mkwrap("screw.Truncate", name)

	if IsWrongCase(name) {
		return wrap(ErrCaseConflict)
	}

	return os.Truncate(name, size)
}

func Readlink(name string) (string, error) {
	debugf("screw.Readlink(%s)", name)
	wrap := mkwrap("screw.Readlink", name)

	if IsWrongCase(name) {
		return "", wrap(os.ErrNotExist)
	}

	return os.Readlink(name)
}

func ReadDir(dirname string) ([]os.FileInfo, error) {
	debugf("screw.ReadDir(%s)", dirname)
	wrap := mkwrap("screw.ReadDir", dirname)

	if IsWrongCase(dirname) {
		return nil, wrap(os.ErrNotExist)
	}

	return ioutil.ReadDir(dirname)
}

func ReadFile(filename string) ([]byte, error) {
	debugf("screw.ReadFile(%s)", filename)
	wrap := mkwrap("screw.ReadFile", filename)

	if IsWrongCase(filename) {
		return nil, wrap(os.ErrNotExist)
	}

	return ioutil.ReadFile(filename)
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	debugf("screw.WriteFile(%s)", filename)
	wrap := mkwrap("screw.WriteFile", filename)

	if IsWrongCase(filename) {
		return wrap(ErrCaseConflict)
	}

	return ioutil.WriteFile(filename, data, perm)
}

func Mkdir(name string, perm os.FileMode) error {
	debugf("screw.Mkdir(%s)", name)
	wrap := mkwrap("screw.Mkdir", name)

	if IsWrongCase(name) {
		return wrap(ErrCaseConflict)
	}

	return os.Mkdir(name, perm)
}

func MkdirAll(name string, perm os.FileMode) error {
	debugf("screw.MkdirAll(%s)", name)
	wrap := mkwrap("screw.MkdirAll", name)

	if IsWrongCase(name) {
		return wrap(ErrCaseConflict)
	}

	return os.MkdirAll(name, perm)
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	debugf("screw.OpenFile(%s, 0x%x, 0o%o)", name, flag, perm)
	wrap := mkwrap("screw.OpenFile", name)

	if IsWrongCase(name) {
		if (flag & os.O_CREATE) > 0 {
			return nil, wrap(ErrCaseConflict)
		} else {
			return nil, wrap(os.ErrNotExist)
		}
	}

	return os.OpenFile(name, flag, perm)
}

func Stat(name string) (os.FileInfo, error) {
	debugf("screw.Stat(%s)", name)
	wrap := mkwrap("screw.Stat", name)

	if IsWrongCase(name) {
		return nil, wrap(os.ErrNotExist)
	}

	return os.Stat(name)
}

func Lstat(name string) (os.FileInfo, error) {
	debugf("screw.Lstat(%s)", name)
	wrap := mkwrap("screw.Lstat", name)

	if IsWrongCase(name) {
		return nil, wrap(os.ErrNotExist)
	}

	return os.Lstat(name)
}

func RemoveAll(name string) error {
	debugf("screw.RemoveAll(%s)", name)
	if IsWrongCase(name) {
		// asked to remove "apricot" but "APRICOT" (or another case variant)
		// exists, consider already removed
		return nil
	}

	return os.RemoveAll(name)
}

func Remove(name string) error {
	debugf("screw.Remove(%s)", name)
	wrap := mkwrap("screw.Remove", name)

	if IsWrongCase(name) {
		// asked to remove "apricot" but "APRICOT" (or another case variant)
		// exists, so, can't remove "apricot" because it doesn't exist
		return wrap(os.ErrNotExist)
	}

	// accepting to try and remove "apricot"
	return os.Remove(name)
}

func Rename(oldpath, newpath string) error {
	debugf("screw.Rename(%s, %s)", oldpath, newpath)
	err := doRename(oldpath, newpath)
	if err != nil {
		return err
	}

	// case-only rename?
	if strings.ToLower(oldpath) == strings.ToLower(newpath) {
		// was it changed properly?
		if TrueBaseName(newpath) != filepath.Base(newpath) {
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

func debugf(f string, arg ...interface{}) {
	if DEBUG {
		line := fmt.Sprintf(f, arg...)
		fmt.Printf("[screw] %s\n", line)
		sneakyLog(line)
	}
}

func mkwrap(op string, path string) func(err error) error {
	return func(err error) error {
		return wrap(err, op, path)
	}
}

func wrap(err error, op string, path string) error {
	if err != nil {
		return &os.PathError{
			Op:   op,
			Path: path,
			Err:  err,
		}
	}
	return err
}
