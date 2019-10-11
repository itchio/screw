package screw_test

import (
	"errors"
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

func ErrorIs(expected error) func(err error) bool {
	return func(actual error) bool {
		return errors.Is(actual, expected)
	}
}

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

func OpCreate(create func(name string) (*os.File, error)) OpFunc {
	return func(name string) (bool, error) {
		f, err := create(name)
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
	ExistsBefore string

	// name of file to pass to operation
	Argument string

	// operation, see OpOpen, OpStat, OpRemove
	Operation OpFunc

	// if true, operation must succeed
	Success bool

	// if non-nil, operation must fail and the returned error's
	// string representation should contain the string representation
	// of this error
	Error func(err error) bool

	// File names that must exist after this call
	ExistsAfter []string

	// if empty, test on all OSes
	OSes []OS
}

func (tc TestCase) AssertValid() {
	if tc.Argument == "" {
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
	return false
}

func listTestCases() []TestCase {
	var testCases []TestCase

	type statVariant struct {
		name string
		op   OpFunc
	}

	osStatVariants := []statVariant{
		{
			name: "os.Stat",
			op:   OpStat(os.Stat),
		},
		{
			name: "os.Lstat",
			op:   OpStat(os.Lstat),
		},
	}

	for _, variant := range osStatVariants {
		testCases = append(testCases, TestCase{
			Name:         variant.name + " nonexistent",
			ExistsBefore: None,
			Argument:     "apricot",
			Operation:    variant.op,
			Error:        os.IsNotExist,
		})
		testCases = append(testCases, TestCase{
			Name:         variant.name + " wrongcase-windows-darwin",
			ExistsBefore: "APRICOT",
			Argument:     "apricot",
			Operation:    variant.op,
			Success:      true,
			OSes:         []OS{Windows, Darwin},
		})
		testCases = append(testCases, TestCase{
			Name:         variant.name + " wrongcase-linux",
			ExistsBefore: "APRICOT",
			Argument:     "apricot",
			Operation:    variant.op,
			Error:        os.IsNotExist,
			OSes:         []OS{Linux},
		})
		testCases = append(testCases, TestCase{
			Name:         variant.name + " rightcase",
			ExistsBefore: "apricot",
			Argument:     "apricot",
			Operation:    variant.op,
			Success:      true,
		})
	}

	screwStatVariants := []statVariant{
		{
			name: "screw.Stat",
			op:   OpStat(screw.Stat),
		},
		{
			name: "screw.Lstat",
			op:   OpStat(screw.Lstat),
		},
	}

	for _, variant := range screwStatVariants {
		testCases = append(testCases, TestCase{
			Name:         variant.name + " nonexistent",
			ExistsBefore: None,
			Argument:     "apricot",
			Operation:    variant.op,
			Error:        os.IsNotExist,
		})
		testCases = append(testCases, TestCase{
			Name:         variant.name + " wrongcase",
			ExistsBefore: "APRICOT",
			Argument:     "apricot",
			Operation:    variant.op,
			Error:        ErrorIs(screw.ErrWrongCase),
			OSes:         []OS{Windows, Darwin},
		})
		testCases = append(testCases, TestCase{
			Name:         variant.name + " rightcase",
			ExistsBefore: "apricot",
			Argument:     "apricot",
			Operation:    variant.op,
			Success:      true,
		})
	}

	//==========================
	// Create
	//==========================

	testCases = append(testCases, TestCase{
		Name:         "os.Create nonexistent",
		ExistsBefore: None,
		Argument:     "apricot",
		Operation:    OpCreate(os.Create),
		Success:      true,
		ExistsAfter:  []string{"apricot"},
	})
	testCases = append(testCases, TestCase{
		Name:         "os.Create wrongcase windows-darwin",
		ExistsBefore: "APRICOT",
		Argument:     "apricot",
		Operation:    OpCreate(os.Create),
		Success:      true,
		ExistsAfter:  []string{"APRICOT"},
		OSes:         []OS{Windows, Darwin},
	})
	testCases = append(testCases, TestCase{
		Name:         "os.Create wrongcase linux",
		ExistsBefore: "APRICOT",
		Argument:     "apricot",
		Operation:    OpCreate(os.Create),
		Success:      true,
		ExistsAfter:  []string{"apricot", "APRICOT"},
		OSes:         []OS{Linux},
	})
	testCases = append(testCases, TestCase{
		Name:         "os.Create rightcase",
		ExistsBefore: "apricot",
		Argument:     "apricot",
		Operation:    OpCreate(os.Create),
		Success:      true,
		ExistsAfter:  []string{"apricot"},
	})

	testCases = append(testCases, TestCase{
		Name:         "screw.Create nonexistent",
		ExistsBefore: None,
		Argument:     "apricot",
		Operation:    OpCreate(screw.Create),
		Success:      true,
		ExistsAfter:  []string{"apricot"},
	})
	testCases = append(testCases, TestCase{
		Name:         "screw.Create wrongcase windows-darwin",
		ExistsBefore: "APRICOT",
		Argument:     "apricot",
		Operation:    OpCreate(screw.Create),
		Error:        ErrorIs(screw.ErrWrongCase),
		ExistsAfter:  []string{"APRICOT"},
		OSes:         []OS{Windows, Darwin},
	})
	testCases = append(testCases, TestCase{
		Name:         "screw.Create wrongcase linux",
		ExistsBefore: "APRICOT",
		Argument:     "apricot",
		Operation:    OpCreate(screw.Create),
		Success:      true,
		ExistsAfter:  []string{"apricot", "APRICOT"},
		OSes:         []OS{Linux},
	})
	testCases = append(testCases, TestCase{
		Name:         "screw.Create rightcase",
		ExistsBefore: "apricot",
		Argument:     "apricot",
		Operation:    OpCreate(screw.Create),
		Success:      true,
		ExistsAfter:  []string{"apricot"},
	})

	return testCases
}

func Test_Semantics(t *testing.T) {
	testCases := listTestCases()

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

			if tc.ExistsBefore != "" {
				f, err := os.Create(filepath.Join(dir, tc.ExistsBefore))
				must(err)
				must(f.Close())
			}

			success, error := tc.Operation(filepath.Join(dir, tc.Argument))

			if tc.Success {
				assert.True(success, "operation should succeed")
			}

			if tc.Error != nil {
				assert.NotNil(error)
				if error != nil {
					assert.True(tc.Error(error))
				}
			}

			for _, ea := range tc.ExistsAfter {
				_, err := os.Stat(filepath.Join(dir, ea))
				assert.NoError(err, "%s must exist after", ea)
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
