package screw

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
)

var DEBUG = os.Getenv("SCREW_DEBUG") == "1"

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
	stackdebugf("screw.Create (%s)", name)
	f, err := openFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
	debugerr(err, "screw.Create (%s)", name)
	return f, err
}

func Open(name string) (*os.File, error) {
	stackdebugf("screw.Open (%s)", name)
	f, err := openFile(name, os.O_RDONLY, 0)
	debugerr(err, "screw.Open (%s)", name)
	return f, err
}

func Symlink(oldname string, newname string) error {
	stackdebugf("screw.Symlink (%s, %s)", oldname, newname)
	err := os.Symlink(oldname, newname)
	debugerr(err, "screw.Symlink (%s, %s)", oldname, newname)
	return err
}

func Truncate(name string, size int64) error {
	stackdebugf("screw.Truncate (%s, %d)", name, size)
	wrap := mkwrap("screw.Truncate", name)

	if IsWrongCase(name) {
		return wrap(ErrCaseConflict)
	}

	err := os.Truncate(name, size)
	debugerr(err, "screw.Truncate (%s, %d)", name, size)
	return err
}

func Readlink(name string) (string, error) {
	stackdebugf("screw.Readlink (%s)", name)
	wrap := mkwrap("screw.Readlink", name)

	if IsWrongCase(name) {
		return "", wrap(os.ErrNotExist)
	}

	s, err := os.Readlink(name)
	debugerr(err, "screw.Readlink (%s)", name)
	return s, err
}

func ReadDir(dirname string) ([]os.FileInfo, error) {
	stackdebugf("screw.ReadDir (%s)", dirname)
	wrap := mkwrap("screw.ReadDir", dirname)

	if IsWrongCase(dirname) {
		return nil, wrap(os.ErrNotExist)
	}

	e, err := ioutil.ReadDir(dirname)
	debugerr(err, "screw.ReadDir (%s)", dirname)
	return e, err
}

func ReadFile(filename string) ([]byte, error) {
	stackdebugf("screw.ReadFile(%s)", filename)
	wrap := mkwrap("screw.ReadFile", filename)

	if IsWrongCase(filename) {
		return nil, wrap(os.ErrNotExist)
	}

	return ioutil.ReadFile(filename)
}

func WriteFile(filename string, data []byte, perm os.FileMode) error {
	stackdebugf("screw.WriteFile(%s)", filename)
	wrap := mkwrap("screw.WriteFile", filename)

	if IsWrongCase(filename) {
		return wrap(ErrCaseConflict)
	}

	return ioutil.WriteFile(filename, data, perm)
}

func Mkdir(name string, perm os.FileMode) error {
	stackdebugf("screw.Mkdir(%s)", name)
	wrap := mkwrap("screw.Mkdir", name)

	if IsWrongCase(name) {
		return wrap(ErrCaseConflict)
	}

	err := os.Mkdir(name, perm)
	debugerr(err, "screw.Mkdir (%s) (0o%o)", name, perm)
	return err
}

func MkdirAll(name string, perm os.FileMode) error {
	stackdebugf("screw.MkdirAll (%s) (0o%o)", name, perm)
	wrap := mkwrap("screw.MkdirAll", name)

	if IsWrongCase(name) {
		return wrap(ErrCaseConflict)
	}

	err := os.MkdirAll(name, perm)
	debugerr(err, "screw.MkdirAll (%s) (0o%o)", name, perm)
	return err
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	stackdebugf("screw.OpenFile (%s) (0x%x) (0o%o)", name, flag, perm)
	f, err := openFile(name, flag, perm)
	debugerr(err, "screw.OpenFile (%s) (0x%x) (0o%o)", name, flag, perm)
	return f, err
}

func openFile(name string, flag int, perm os.FileMode) (*os.File, error) {
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
	stackdebugf("screw.Stat (%s)", name)
	wrap := mkwrap("screw.Stat", name)

	if IsWrongCase(name) {
		return nil, wrap(os.ErrNotExist)
	}

	s, err := os.Stat(name)
	debugerr(err, "screw.Stat (%s)", name)
	return s, err
}

func Lstat(name string) (os.FileInfo, error) {
	stackdebugf("screw.Lstat (%s)", name)
	wrap := mkwrap("screw.Lstat", name)

	if IsWrongCase(name) {
		return nil, wrap(os.ErrNotExist)
	}

	s, err := os.Lstat(name)
	debugerr(err, "screw.Lstat (%s)", name)
	return s, err
}

func RemoveAll(name string) error {
	stackdebugf("screw.RemoveAll (%s)", name)
	if IsWrongCase(name) {
		// asked to remove "apricot" but "APRICOT" (or another case variant)
		// exists, consider already removed
		return nil
	}

	err := os.RemoveAll(name)
	debugerr(err, "screw.RemoveAll (%s)", name)
	return err
}

func Remove(name string) error {
	debugf("screw.Remove (%s)", name)
	wrap := mkwrap("screw.Remove", name)

	if IsWrongCase(name) {
		// asked to remove "apricot" but "APRICOT" (or another case variant)
		// exists, so, can't remove "apricot" because it doesn't exist
		return wrap(os.ErrNotExist)
	}

	// accepting to try and remove "apricot"
	err := os.Remove(name)
	debugerr(err, "screw.Remove (%s)", name)
	return err
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

func stackdebugf(f string, arg ...interface{}) {
	debugfex(true, f, arg...)
}

func debugf(f string, arg ...interface{}) {
	debugfex(false, f, arg...)
}

func debugerr(err error, f string, arg ...interface{}) {
	if !DEBUG {
		return
	}

	line := fmt.Sprintf(f, arg...)
	if err == nil {
		debugf("[CHEERS] %s", line)
	} else {
		debugf("[SORROW] %s: %+v", line, err)
	}
}

func debugfex(stack bool, f string, arg ...interface{}) {
	if DEBUG {
		line := fmt.Sprintf(f, arg...)
		fmt.Printf("\n\n")
		fmt.Printf("=============[screw]=============\n")
		fmt.Printf("%s\n", line)
		if stack {
			stackLines := strings.Split(string(debug.Stack()), "\n")
			for _, stackLine := range stackLines {
				fmt.Printf("     -> %s\n", stackLine)
			}
		}
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
