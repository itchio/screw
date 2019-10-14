// +build linux

package screw

import "os"

// If `path` exists, and
func TrueBaseName(path string) string {
	_, err := os.Stat(path)
	if err != nil {
		return "", os.ErrNotExist
	}
	return path, nil
}
