package screw_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/windows"
)

func Test_Forbidden_RemoveWhileOpen(t *testing.T) {
	assert := assert.New(t)

	if runtime.GOOS != "windows" {
		t.Skip()
	}

	tmpdir, err := ioutil.TempDir("", "")
	must(err)

	name := filepath.Join(tmpdir, "hi.txt")

	must(ioutil.WriteFile(name, []byte("Hello"), 0o644))

	f, err := os.Open(name)
	must(err)
	defer f.Close()

	err = os.Remove(name)
	t.Logf("Remove error: %v", err)
	assert.Error(err)

	pe, ok := err.(*os.PathError)
	assert.True(ok)
	if !ok {
		return
	}
	assert.Equal(windows.ERROR_SHARING_VIOLATION, pe.Err)
}

func Test_Forbidden_RemoveParentWhileOpen(t *testing.T) {
	assert := assert.New(t)

	if runtime.GOOS != "windows" {
		t.Skip()
	}

	tmpdir, err := ioutil.TempDir("", "")
	must(err)

	parent := filepath.Join(tmpdir, "parent")
	must(os.MkdirAll(parent, 0o755))

	name := filepath.Join(parent, "hi.txt")
	must(ioutil.WriteFile(name, []byte("Hello"), 0o644))

	f, err := os.Open(name)
	must(err)
	defer f.Close()

	err = os.RemoveAll(parent)
	t.Logf("RemoveAll error: %v", err)
	assert.Error(err)

	pe, ok := err.(*os.PathError)
	assert.True(ok)
	if !ok {
		return
	}
	assert.Equal(windows.ERROR_SHARING_VIOLATION, pe.Err)
}
