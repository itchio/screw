package screw_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/itchio/screw"
	"github.com/stretchr/testify/assert"
)

// Note: on macOS, these tests assume we're running on
// a case-sensitive filesystem like HFS+ or APFS

func Test_Open(t *testing.T) {
	if runtime.GOOS != "windows" {
		return
	}

	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "screw-test-open")
	must(err)

	defer os.RemoveAll(tmpDir)

	must(os.Mkdir(filepath.Join(tmpDir, "foo"), os.FileMode(755)))

	f, err := os.Create(filepath.Join(tmpDir, "foo", "bar.txt"))
	must(err)
	must(f.Close())

	canOpen := func(open func(name string) (*os.File, error), name string) {
		f, err := open(name)
		must(err)
		must(f.Close())
	}

	canStat := func(stat func(name string) (os.FileInfo, error), name string) {
		_, err := stat(name)
		must(err)
	}

	wontOpen := func(open func(name string) (*os.File, error), name string) {
		_, err := open(name)
		assert.Error(err)
		if err == nil {
			t.FailNow()
		}
	}

	wontStat := func(stat func(name string) (os.FileInfo, error), name string) {
		_, err := stat(name)
		assert.Error(err)
		if err == nil {
			t.FailNow()
		}
	}

	joinPath := func(a ...string) string {
		return strings.Join(a, `\`)
	}

	paths := []string{
		joinPath(tmpDir, "foo", "bar"),
		joinPath(tmpDir, "foo", "BAR"),
		joinPath(tmpDir, "FOO", "bar"),
	}

	// Windows can operate on any of these paths
	for _, path := range paths {
		canOpen(os.Open, path)
		canStat(os.Stat, path)
		canStat(os.Lstat, path)
	}

	// Screw wants the right one
	for i, path := range paths {
		if i == 0 {
			canOpen(screw.Open, path)
			canStat(screw.Stat, path)
			canStat(screw.Lstat, path)
		} else {
			wontOpen(screw.Open, path)
			wontStat(screw.Stat, path)
			wontStat(screw.Lstat, path)
		}
	}
}

func Test_RenameCase(t *testing.T) {
	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "screw-test-rename")
	must(err)

	defer os.RemoveAll(tmpDir)

	f, err := os.Create(filepath.Join(tmpDir, "foobar"))
	must(err)
	must(f.Close())

	assert.True(screw.IsActualCasing(filepath.Join(tmpDir, "foobar")))
	must(screw.Rename(filepath.Join(tmpDir, "foobar"), filepath.Join(tmpDir, "Foobar")))
	assert.True(screw.IsActualCasing(filepath.Join(tmpDir, "Foobar")))
}

func Test_RenameLocked(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "screw-test-rename")
	must(err)

	defer os.RemoveAll(tmpDir)

	f, err := os.Create(filepath.Join(tmpDir, "foobar"))
	must(err)

	go func() {
		time.Sleep(2 * time.Second)
		f.Close()
	}()

	must(screw.Rename(filepath.Join(tmpDir, "foobar"), filepath.Join(tmpDir, "something-else")))
}

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
