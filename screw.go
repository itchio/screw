package screw

import (
	"errors"
	"os"
)

var DEBUG = os.Getenv("DEBUG_SCREW") == "1"

var (
	ErrWrongCase = errors.New("a file with a different case exists on disk")
)

func Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

func Create(name string) (*os.File, error) {
	return OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
}

func Open(name string) (*os.File, error) {
	return OpenFile(name, os.O_RDONLY, 0)
}
