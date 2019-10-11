package screw_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/itchio/screw"
	"github.com/stretchr/testify/assert"
)

type OpFunc func(name string) (bool, error)

func OpOpen(open func(name string) (*os.File, error)) OpFunc {
	return func(name string) (bool, error) {
		f, err := open(name)
		if err != nil {
			return false, err
		}

		f.Close()
		return true, nil
	}
}

func OpStat(stat func(name string) (os.FileInfo, error)) OpFunc {
	return func(name string) (bool, error) {
		_, err := stat(name)
		if err != nil {
			return false, err
		}
		return true, nil
	}
}

func OpRemove(remove func(name string) error) OpFunc {
	return func(name string) (bool, error) {
		err := remove(name)
		if err != nil {
			return false, err
		}
		return true, nil
	}
}

type OS string

const (
	Windows OS = "windows"
	Linux   OS = "linux"
	Darwin  OS = "darwin"
)

const None = ""

type TestCase struct {
	// name of test
	Name string

	// name of file to create. if empty, no file is created
	CreateFile string

	// name of file to pass to operation
	PassFile string

	// operation, see OpOpen, OpStat, OpRemove
	Operation OpFunc

	// if true, operation must succeed
	Success bool

	// if non-nil, operation must fail and the returned error's
	// string representation should contain the string representation
	// of this error
	Error func(err error) bool

	// if empty, test on all OSes
	OSes []OS
}

func (tc TestCase) AssertValid() {
	if tc.PassFile == "" {
		panic("invalid test: empty PassFile")
	}

	if tc.Success && tc.Error != nil {
		panic("invalid test: both Success and Error specified")
	}
}

func (tc TestCase) ShouldRun(t *testing.T) bool {
	if len(tc.OSes) == 0 {
		return true
	}

	for _, os := range tc.OSes {
		if string(os) == runtime.GOOS {
			return true
		}
	}

	t.Logf("We're on (%s), skipping (%s)", runtime.GOOS, tc.Name)
	return false
}

var testCases = []TestCase{
	TestCase{
		Name:       "os.Stat, non-existent file",
		CreateFile: None,
		PassFile:   "apricot",
		Operation:  OpStat(os.Stat),
		Error:      os.IsNotExist,
	},
	TestCase{
		Name:       "screw.Stat, non-existent file",
		CreateFile: None,
		PassFile:   "apricot",
		Operation:  OpStat(screw.Stat),
		Error:      os.IsNotExist,
	},
	TestCase{
		Name:       "os.Lstat, non-existent file",
		CreateFile: None,
		PassFile:   "apricot",
		Operation:  OpStat(os.Lstat),
		Error:      os.IsNotExist,
	},
	TestCase{
		Name:       "screw.Lstat, non-existent file",
		CreateFile: None,
		PassFile:   "apricot",
		Operation:  OpStat(screw.Lstat),
		Error:      os.IsNotExist,
	},
}

func Test_Semantics(t *testing.T) {
	for _, tc := range testCases {
		tc.AssertValid()
		if !tc.ShouldRun(t) {
			continue
		}

		t.Run(tc.Name, func(t *testing.T) {
			assert := assert.New(t)

			dir, err := ioutil.TempDir("", "screw-tests")
			must(err)
			defer os.RemoveAll(dir)

			if tc.CreateFile != "" {
				f, err := os.Create(filepath.Join(dir, tc.CreateFile))
				must(err)
				must(f.Close())
			}

			success, error := tc.Operation(tc.PassFile)

			if tc.Success {
				assert.True(success, "operation should succeed")
			}

			if tc.Error != nil {
				assert.NotNil(error)
				if error != nil {
					assert.True(tc.Error(error))
				}
			}
		})
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
		time.Sleep(200 * time.Millisecond)
		f.Close()
	}()

	must(screw.Rename(filepath.Join(tmpDir, "foobar"), filepath.Join(tmpDir, "something-else")))
}

func must(err error) {
	if err != nil {
		panic(fmt.Sprintf("%+v", err))
	}
}
