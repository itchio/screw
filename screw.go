package screw

import "os"

var DEBUG = os.Getenv("DEBUG_SCREW") == "1"

func Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

func Create(name string) (*os.File, error) {
	return os.Create(name)
}

func Open(name string) (*os.File, error) {
	return OpenFile(name, os.O_RDONLY, 0)
}
