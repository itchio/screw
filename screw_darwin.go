//go:build darwin

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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"unsafe"
)

// osRename is a test seam so darwin-only tests can deterministically exercise
// fallback and rollback branches in doRename without relying on filesystem quirks.
var osRename = os.Rename

func sneakyLog(line string) {
	// nothing
}

// If `path` exists, and
func TrueBaseName(path string) string {
	cInputPath := C.CString(path)
	defer C.free(unsafe.Pointer(cInputPath))

	cPath := C.GetCanonicalPath(cInputPath)
	if uintptr(unsafe.Pointer(cPath)) == 0 {
		return ""
	}
	defer C.free(unsafe.Pointer(cPath))

	actualPath := C.GoString(cPath)
	return filepath.Base(actualPath)
}

func doRename(oldpath, newpath string) error {
	err := osRename(oldpath, newpath)
	if err != nil {
		// on macOS, renaming "foo" to "FOO" is fine,
		// but renaming "foo/" to "FOO/" fails with "file exists"
		if os.IsExist(err) {
			originalErr := err

			tmppath := fmt.Sprintf("%s__rename_pid%d", oldpath, os.Getpid())

			err = osRename(oldpath, tmppath)
			if err != nil {
				return originalErr
			}
			err = osRename(tmppath, newpath)
			if err != nil {
				// attempt to rollback, ignore returned error,
				// because what else can we do at this point?
				rollbackErr := osRename(tmppath, oldpath)
				if rollbackErr != nil {
					return errors.Join(err, rollbackErr)
				}
				return err
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
