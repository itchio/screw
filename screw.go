package screw

import (
	"errors"
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
