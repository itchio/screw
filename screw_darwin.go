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
	"path/filepath"
	"unsafe"
)

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
