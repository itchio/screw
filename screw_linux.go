// +build linux

package screw

import "os"

// If `path` exists, and
func TrueBaseName(path string) string {
	stats, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return stats.Name()
}

func doRename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}
