package screw

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

var DEBUG = os.Getenv("DEBUG_SCREW") == "1"

var (
	ErrCaseConflict = errors.New("a file with a different case already exists on disk")
)

func Create(name string) (*os.File, error) {
	return OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o666)
}

func Open(name string) (*os.File, error) {
	return OpenFile(name, os.O_RDONLY, 0)
}

func Symlink(oldname string, newname string) error {
	return os.Symlink(oldname, newname)
}

func ReadDir(dirname string) ([]os.FileInfo, error) {
	wrap := mkwrap("screw.ReadDir", dirname)

	isActual, err := IsActualCasing(dirname)
	if err != nil {
		return nil, wrap(err)
	}

	if !isActual {
		return nil, wrap(os.ErrNotExist)
	}

	entries, err := ioutil.ReadDir(dirname)
	if err != nil {
		return nil, wrap(err)
	}
	return entries, nil
}

func ReadFile(filename string) ([]byte, error) {
	wrap := mkwrap("screw.ReadFile", filename)

	isActual, err := IsActualCasing(filename)
	if err != nil {
		return nil, wrap(err)
	}

	if !isActual {
		return nil, wrap(os.ErrNotExist)
	}

	bs, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, wrap(err)
	}
	return bs, nil
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
