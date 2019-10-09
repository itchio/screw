package screw_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/itchio/screw"
	"github.com/stretchr/testify/assert"
)

func Test_Open(t *testing.T) {
	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "screw-tests")
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

	ensureNop := func(f *os.File, err error) {
		assert.Error(err)
		assert.Nil(f)
	}

	ensureYay(os.Open(tmpDir + `\foo\bar.txt`))
	ensureYay(os.Open(tmpDir + `\foo\BAR.txt`))
	ensureYay(os.Open(tmpDir + `\FOO\bar.txt`))
	ensureYay(os.Open(tmpDir + `\FOO\..\foo\bar.txt`))

	ensureYay(screw.Open(tmpDir + `\foo\bar.txt`))
	ensureNop(screw.Open(tmpDir + `\foo\BAR.txt`))
	ensureNop(screw.Open(tmpDir + `\FOO\bar.txt`))
	ensureYay(screw.Open(tmpDir + `\FOO\..\foo\bar.txt`))
}

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
