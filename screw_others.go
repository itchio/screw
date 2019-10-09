// +build !windows

package screw

import "os"

func Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func OpenFile(name string, flag int, perm os.FileMode) (*os.File, error) {
	return os.OpenFile(name, flag, perm)
}

func IsActualCasing(path string) (bool, error) {
	_, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return true, nil
}
