// +build darwin linux

package screw

import "os"

func Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func MkdirAll(name string, perm os.FileMode) error {
	return os.MkdirAll(name, perm)
}

func Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func Lstat(name string) (os.FileInfo, error) {
	return os.Lstat(name)
}

func RemoveAll(name string) error {
	return os.RemoveAll(name)
}

func Remove(name string) error {
	return os.Remove(name)
}
