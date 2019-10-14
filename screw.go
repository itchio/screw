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
	return OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
}

func Open(name string) (*os.File, error) {
	return OpenFile(name, os.O_RDONLY, 0)
}

func Symlink(oldname string, newname string) error {
	return os.Symlink(oldname, newname)
}

func Readlink(name string) (string, error) {
	wrap := mkwrap("screw.Readlink", name)

	if IsWrongCase(name) {
		return "", wrap(os.ErrNotExist)
	}

	return os.Readlink(name)
}

func ReadDir(dirname string) ([]os.FileInfo, error) {
	wrap := mkwrap("screw.ReadDir", dirname)

	if IsWrongCase(dirname) {
		return nil, wrap(os.ErrNotExist)
	}

	return ioutil.ReadDir(dirname)
}

func ReadFile(filename string) ([]byte, error) {
	wrap := mkwrap("screw.ReadFile", filename)

	if IsWrongCase(filename) {
		return nil, wrap(os.ErrNotExist)
	}

	return ioutil.ReadFile(filename)
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	wrap := mkwrap("screw.WriteFile", filename)

	if IsWrongCase(filename) {
		return wrap(ErrCaseConflict)
	}

	return ioutil.WriteFile(filename, data, perm)
}

func Mkdir(name string, perm os.FileMode) error {
	wrap := mkwrap("screw.Mkdir", name)

	if IsWrongCase(name) {
		return wrap(ErrCaseConflict)
	}

	return os.Mkdir(name, perm)
}

func MkdirAll(name string, perm os.FileMode) error {
	wrap := mkwrap("screw.MkdirAll", name)

	if IsWrongCase(name) {
		return wrap(ErrCaseConflict)
	}

	return os.MkdirAll(name, perm)
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
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
	wrap := mkwrap("screw.Stat", name)

	if IsWrongCase(name) {
		return nil, wrap(os.ErrNotExist)
	}

	return os.Stat(name)
}

func Lstat(name string) (os.FileInfo, error) {
	wrap := mkwrap("screw.Lstat", name)

	if IsWrongCase(name) {
		return nil, wrap(os.ErrNotExist)
	}

	return os.Lstat(name)
}

func RemoveAll(name string) error {
	if IsWrongCase(name) {
		// asked to remove "apricot" but "APRICOT" (or another case variant)
		// exists, consider already removed
		return nil
	}

	return os.RemoveAll(name)
}

func Remove(name string) error {
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
		fmt.Printf("[screw] %s\n", fmt.Sprintf(f, arg...))
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
