//go:build darwin
// +build darwin

package screw

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>
#include <stdlib.h>

char *GetCanonicalPath(char *cInputPath) {
	NSString *inputPath = [NSString stringWithUTF8String:cInputPath];
	if (!inputPath) {
		return NULL;
	}
	NSString *canonicalPath = [[[NSURL fileURLWithPath:inputPath] fileReferenceURL] path];
	if (!canonicalPath) {
		return NULL;
	}

	const char *tempString = [canonicalPath UTF8String];
	char *ret = malloc(strlen(tempString) + 1);
	memcpy(ret, tempString, strlen(tempString) + 1);
	return ret;
}
*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

func sneakyLog(line string) {
	// nothing
}

// If `path` exists, and
func TrueBaseName(path string) string {
	cPath := C.GetCanonicalPath(C.CString(path))
	if uintptr(unsafe.Pointer(cPath)) == 0 {
		return ""
	}
	defer C.free(unsafe.Pointer(cPath))

	actualPath := C.GoString(cPath)
	return filepath.Base(actualPath)
}

func doRename(oldpath, newpath string) error {
	err := os.Rename(oldpath, newpath)
	if err != nil {
		// on macOS, renaming "foo" to "FOO" is fine,
		// but renaming "foo/" to "FOO/" fails with "file exists"
		if os.IsExist(err) {
			originalErr := err

			tmppath := fmt.Sprintf("%s__rename_pid%d", oldpath, os.Getpid())

			err = os.Rename(oldpath, tmppath)
			if err != nil {
				return originalErr
			}
			err = os.Rename(tmppath, newpath)
			if err != nil {
				// attempt to rollback, ignore returned error,
				// because what else can we do at this point?
				_ = os.Rename(tmppath, oldpath)
			}
			// two-stage rename worked!
			return nil
		}
		return err
	}
	return nil
}

func IsCaseInsensitiveFS() bool {
	return true
}
