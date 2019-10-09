package screw_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/itchio/screw"
	"github.com/stretchr/testify/assert"
)

func Test_Open(t *testing.T) {
	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "screw-test-open")
	must(err)

	defer os.RemoveAll(tmpDir)

	must(os.Mkdir(filepath.Join(tmpDir, "foo"), os.FileMode(755)))

	f, err := os.Create(filepath.Join(tmpDir, "foo", "bar.txt"))
	must(err)
	must(f.Close())

	ensureYay := func(f *os.File, err error) {
		assert.NotNil(f)
		assert.NoError(err)
		must(f.Close())
	}

	ensureNah := func(f *os.File, err error) {
		assert.Error(err)
		assert.Nil(f)
	}

	// os
	ensureYay(os.Open(tmpDir + `\foo\bar.txt`))
	ensureYay(os.Open(tmpDir + `\FOO\..\foo\bar.txt`))
	if runtime.GOOS == "windows" {
		ensureYay(os.Open(tmpDir + `\foo\BAR.txt`))
		ensureYay(os.Open(tmpDir + `\FOO\bar.txt`))
	} else {
		ensureNah(os.Open(tmpDir + `\foo\BAR.txt`))
		ensureNah(os.Open(tmpDir + `\FOO\bar.txt`))
	}

	// screw
	ensureYay(screw.Open(tmpDir + `\foo\bar.txt`))
	ensureYay(screw.Open(tmpDir + `\FOO\..\foo\bar.txt`))

	ensureNah(screw.Open(tmpDir + `\foo\BAR.txt`))
	ensureNah(screw.Open(tmpDir + `\FOO\bar.txt`))
}

func Test_Rename(t *testing.T) {
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

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
