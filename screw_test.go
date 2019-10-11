package screw_test

import (
	"errors"
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

func OpReadFile(readfile func(name string) ([]byte, error)) OpFunc {
	return func(name string) (bool, error) {
		_, err := readfile(name)
		if err != nil {
			return false, err
		}
		return true, nil
	}
}

func OpReadDir(readdir func(name string) ([]os.FileInfo, error)) OpFunc {
	return func(name string) (bool, error) {
		_, err := readdir(name)
		if err != nil {
			return false, err
		}
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

func OpMkdir(mkdir func(name string, perm os.FileMode) error) OpFunc {
	return func(name string) (bool, error) {
		err := mkdir(name, 0o755)
		if err != nil {
			return false, err
		}

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

type FSKind string

const (
	FSAny             FSKind = ""
	FSCaseSensitive   FSKind = "sensitive"
	FSCaseInsensitive FSKind = "insensitive"
)

func (fsk FSKind) String() string {
	switch fsk {
	case FSCaseSensitive:
		return "sensitive-fs"
	case FSCaseInsensitive:
		return "insensitive-fs"
	default:
		return "any-fs"
	}
}

type TestCase struct {
	// name of test
	Name string

	// Dirs to create before this call
	DirsBefore []string

	// Files to create before this call
	FilesBefore []string

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

	// Files that must exist after this call
	// Note that existence will be checked ignoring case on Windows & Darwin
	FilesAfter []string

	// Dirs that must exist after this call
	// Note that existence will be checked ignoring case on Windows & Darwin
	DirsAfter []string

	// Files/dirs that must *not* exist after this call
	AbsentAfter []string

	// if empty, test on all OSes
	FSKind FSKind
}

func (tc TestCase) AssertValid() {
	if tc.Argument == "" {
		panic("invalid test: empty PassFile")
	}

	if tc.Success && tc.Error != nil {
		panic("invalid test: both Success and Error specified")
	}

	if !tc.Success && tc.Error == nil {
		panic("invalid test: neither Success nor Error specified")
	}
}

func (tc TestCase) ShouldRun(t *testing.T) bool {
	switch tc.FSKind {
	case FSCaseInsensitive:
		return runtime.GOOS == "windows" || runtime.GOOS == "darwin"
	case FSCaseSensitive:
		return runtime.GOOS == "linux"
	default:
		return true
	}
}

func listTestCases() []TestCase {
	var testCases []TestCase

	type opVariant struct {
		name string
		op   OpFunc
	}

	//==========================
	// Stat, Lstat, Open
	//==========================

	osVariants := []opVariant{
		{
			name: "os.Stat",
			op:   OpStat(os.Stat),
		},
		{
			name: "os.Lstat",
			op:   OpStat(os.Lstat),
		},
		{
			name: "os.Open", // not technically a stat, but..
			op:   OpOpen(os.Open),
		},
	}

	for _, variant := range osVariants {
		testCases = append(testCases, TestCase{
			Name:      variant.name + "/nonexistent",
			Argument:  "apricot",
			Operation: variant.op,
			Error:     os.IsNotExist,
		})
		testCases = append(testCases, TestCase{
			Name:        variant.name + "/mixedcase",
			FilesBefore: []string{"APRICOT"},
			Argument:    "apricot",
			Operation:   variant.op,
			Success:     true,

			FSKind: FSCaseInsensitive,
		})
		testCases = append(testCases, TestCase{
			Name:        variant.name + "/wrongcase",
			FilesBefore: []string{"APRICOT"},
			Argument:    "apricot",
			Operation:   variant.op,
			Error:       os.IsNotExist,

			FSKind: FSCaseSensitive,
		})
		testCases = append(testCases, TestCase{
			Name:        variant.name + "/rightcase",
			FilesBefore: []string{"apricot"},
			Argument:    "apricot",
			Operation:   variant.op,
			Success:     true,
		})
	}

	screwVariants := []opVariant{
		{
			name: "screw.Stat",
			op:   OpStat(screw.Stat),
		},
		{
			name: "screw.Lstat",
			op:   OpStat(screw.Lstat),
		},
		{
			name: "screw.Open",
			op:   OpOpen(screw.Open),
		},
	}

	for _, variant := range screwVariants {
		testCases = append(testCases, TestCase{
			Name:      variant.name + "/nonexistent",
			Argument:  "apricot",
			Operation: variant.op,
			Error:     os.IsNotExist,
		})
		testCases = append(testCases, TestCase{
			Name:        variant.name + "/mixedcase",
			FilesBefore: []string{"APRICOT"},
			Argument:    "apricot",
			Operation:   variant.op,
			Error:       os.IsNotExist,

			FSKind: FSCaseInsensitive,
		})
		testCases = append(testCases, TestCase{
			Name:        variant.name + "/wrongcase",
			FilesBefore: []string{"APRICOT"},
			Argument:    "apricot",
			Operation:   variant.op,
			Error:       os.IsNotExist,

			FSKind: FSCaseSensitive,
		})
		testCases = append(testCases, TestCase{
			Name:        variant.name + "/rightcase",
			FilesBefore: []string{"apricot"},
			Argument:    "apricot",
			Operation:   variant.op,
			Success:     true,
		})
	}

	//==========================
	// ReadFile
	//==========================

	testCases = append(testCases, TestCase{
		Name:      "ioutil.ReadFile/nonexistent",
		Argument:  "apricot",
		Operation: OpReadFile(ioutil.ReadFile),
		Error:     os.IsNotExist,
	})

	testCases = append(testCases, TestCase{
		Name:        "ioutil.ReadFile/mixedcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpReadFile(ioutil.ReadFile),
		Success:     true,

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:        "ioutil.ReadFile/wrongcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpReadFile(ioutil.ReadFile),
		Error:       os.IsNotExist,

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:        "ioutil.ReadFile/rightcase",
		FilesBefore: []string{"apricot"},
		Argument:    "apricot",
		Operation:   OpReadFile(ioutil.ReadFile),
		Success:     true,
	})

	testCases = append(testCases, TestCase{
		Name:      "screw.ReadFile/nonexistent",
		Argument:  "apricot",
		Operation: OpReadFile(screw.ReadFile),
		Error:     os.IsNotExist,
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.ReadFile/wrongcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpReadFile(screw.ReadFile),
		Error:       os.IsNotExist,
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.ReadFile/rightcase",
		FilesBefore: []string{"apricot"},
		Argument:    "apricot",
		Operation:   OpReadFile(screw.ReadFile),
		Success:     true,
	})

	//==========================
	// ReadDir
	//==========================

	testCases = append(testCases, TestCase{
		Name:      "ioutil.ReadDir/nonexistent",
		Argument:  "apricot",
		Operation: OpReadDir(ioutil.ReadDir),
		Error:     os.IsNotExist,
	})

	testCases = append(testCases, TestCase{
		Name:       "ioutil.ReadDir/mixedcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpReadDir(ioutil.ReadDir),
		Success:    true,

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "ioutil.ReadDir/wrongcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpReadDir(ioutil.ReadDir),
		Error:      os.IsNotExist,

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "ioutil.ReadDir/rightcase",
		DirsBefore: []string{"apricot"},
		Argument:   "apricot",
		Operation:  OpReadDir(ioutil.ReadDir),
		Success:    true,
	})

	testCases = append(testCases, TestCase{
		Name:      "screw.ReadDir/nonexistent",
		Argument:  "apricot",
		Operation: OpReadDir(screw.ReadDir),
		Error:     os.IsNotExist,
	})

	testCases = append(testCases, TestCase{
		Name:       "screw.ReadDir/wrongcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpReadDir(screw.ReadDir),
		Error:      os.IsNotExist,
	})

	testCases = append(testCases, TestCase{
		Name:       "screw.ReadDir/rightcase",
		DirsBefore: []string{"apricot"},
		Argument:   "apricot",
		Operation:  OpReadDir(screw.ReadDir),
		Success:    true,
	})

	//==========================
	// Create
	//==========================

	testCases = append(testCases, TestCase{
		Name:       "os.Create/nonexistent",
		Argument:   "apricot",
		Operation:  OpCreate(os.Create),
		Success:    true,
		FilesAfter: []string{"apricot"},
	})
	testCases = append(testCases, TestCase{
		Name:        "os.Create/mixedcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpCreate(os.Create),
		Success:     true,
		FilesAfter:  []string{"APRICOT"},

		FSKind: FSCaseInsensitive,
	})
	testCases = append(testCases, TestCase{
		Name:        "os.Create/wrongcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpCreate(os.Create),
		Success:     true,
		FilesAfter:  []string{"apricot", "APRICOT"},

		FSKind: FSCaseSensitive,
	})
	testCases = append(testCases, TestCase{
		Name:        "os.Create/rightcase",
		FilesBefore: []string{"apricot"},
		Argument:    "apricot",
		Operation:   OpCreate(os.Create),
		Success:     true,
		FilesAfter:  []string{"apricot"},
	})
	testCases = append(testCases, TestCase{
		Name:       "screw.Create/nonexistent",
		Argument:   "apricot",
		Operation:  OpCreate(screw.Create),
		Success:    true,
		FilesAfter: []string{"apricot"},
	})
	testCases = append(testCases, TestCase{
		Name:        "screw.Create/mixedcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpCreate(screw.Create),
		Error:       ErrorIs(screw.ErrCaseConflict),
		FilesAfter:  []string{"APRICOT"},

		FSKind: FSCaseInsensitive,
	})
	testCases = append(testCases, TestCase{
		Name:        "screw.Create/wrongcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpCreate(screw.Create),
		Success:     true,
		FilesAfter:  []string{"apricot", "APRICOT"},

		FSKind: FSCaseSensitive,
	})
	testCases = append(testCases, TestCase{
		Name:        "screw.Create/rightcase",
		FilesBefore: []string{"apricot"},
		Argument:    "apricot",
		Operation:   OpCreate(screw.Create),
		Success:     true,
		FilesAfter:  []string{"apricot"},
	})

	//==========================
	// Remove
	//==========================

	testCases = append(testCases, TestCase{
		Name:      "os.Remove/nonexistent",
		Argument:  "apricot",
		Operation: OpRemove(os.Remove),
		Error:     os.IsNotExist,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.Remove/mixedcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpRemove(os.Remove),
		Success:     true,
		AbsentAfter: []string{"APRICOT"},

		FSKind: FSCaseInsensitive,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.Remove/wrongcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpRemove(os.Remove),
		Error:       os.IsNotExist,

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.Remove/rightcase",
		FilesBefore: []string{"apricot"},
		Argument:    "apricot",
		Operation:   OpRemove(os.Remove),
		Success:     true,
		AbsentAfter: []string{"apricot"},
	})

	testCases = append(testCases, TestCase{
		Name:      "screw.Remove/nonexistent",
		Argument:  "apricot",
		Operation: OpRemove(screw.Remove),
		Error:     os.IsNotExist,
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.Remove/wrongcase",
		FilesBefore: []string{"APRICOT"},
		Argument:    "apricot",
		Operation:   OpRemove(screw.Remove),
		Error:       os.IsNotExist,
		FilesAfter:  []string{"APRICOT"},
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.Remove/rightcase",
		FilesBefore: []string{"apricot"},
		Argument:    "apricot",
		Operation:   OpRemove(screw.Remove),
		Success:     true,
		AbsentAfter: []string{"apricot"},
	})

	//==========================
	// RemoveAll
	//==========================

	testCases = append(testCases, TestCase{
		Name:      "os.RemoveAll/nonexistent",
		Argument:  "apricot",
		Operation: OpRemove(os.RemoveAll),
		Success:   true,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.RemoveAll/mixedcase",
		FilesBefore: []string{"APRICOT/README"},
		Argument:    "apricot",
		Operation:   OpRemove(os.RemoveAll),
		AbsentAfter: []string{"APRICOT/README", "APRICOT"},
		Success:     true,

		FSKind: FSCaseInsensitive,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.RemoveAll/wrongcase",
		FilesBefore: []string{"APRICOT/README"},
		Argument:    "apricot",
		Operation:   OpRemove(os.RemoveAll),
		Success:     true,
		FilesAfter:  []string{"APRICOT/README"},

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.RemoveAll/rightcase",
		FilesBefore: []string{"apricot/README"},
		Argument:    "apricot",
		Operation:   OpRemove(os.RemoveAll),
		AbsentAfter: []string{"apricot/README", "apricot"},
		Success:     true,
	})

	testCases = append(testCases, TestCase{
		Name:      "screw.RemoveAll/nonexistent",
		Argument:  "apricot",
		Operation: OpRemove(screw.RemoveAll),
		Success:   true,
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.RemoveAll/wrongcase",
		FilesBefore: []string{"APRICOT/README"},
		Argument:    "apricot",
		Operation:   OpRemove(screw.RemoveAll),
		Success:     true,
		FilesAfter:  []string{"APRICOT/README"},
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.RemoveAll/rightcase",
		FilesBefore: []string{"apricot/README"},
		Argument:    "apricot",
		Operation:   OpRemove(screw.RemoveAll),
		AbsentAfter: []string{"apricot/README", "apricot"},
		Success:     true,
	})

	//==========================
	// Mkdir
	//==========================

	testCases = append(testCases, TestCase{
		Name:      "os.Mkdir/nonexistent",
		Argument:  "apricot",
		Operation: OpMkdir(os.Mkdir),
		Success:   true,
	})

	testCases = append(testCases, TestCase{
		Name:       "os.Mkdir/mixedcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpMkdir(os.Mkdir),
		Error:      os.IsExist,

		FSKind: FSCaseInsensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "os.Mkdir/wrongcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpMkdir(os.Mkdir),
		Success:    true,

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "os.Mkdir/rightcase",
		DirsBefore: []string{"apricot"},
		Argument:   "apricot",
		Operation:  OpMkdir(os.Mkdir),
		Error:      os.IsExist,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.Mkdir/nonexistentparent",
		Argument:    "foo/bar",
		Operation:   OpMkdir(os.Mkdir),
		Error:       os.IsNotExist,
		AbsentAfter: []string{"foo", "foo/bar"},
	})

	testCases = append(testCases, TestCase{
		Name:       "os.Mkdir/mixedcaseparent",
		DirsBefore: []string{"FOO"},
		Argument:   "foo/bar",
		Operation:  OpMkdir(os.Mkdir),
		Success:    true,
		DirsAfter:  []string{"FOO", "FOO/bar"},

		FSKind: FSCaseInsensitive,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.Mkdir/wrongcaseparent",
		DirsBefore:  []string{"FOO"},
		Argument:    "foo/bar",
		Operation:   OpMkdir(os.Mkdir),
		Error:       os.IsNotExist,
		AbsentAfter: []string{"foo", "foo/bar"},

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "os.Mkdir/rightcaseparent",
		DirsBefore: []string{"foo"},
		Argument:   "foo/bar",
		Operation:  OpMkdir(os.Mkdir),
		DirsAfter:  []string{"foo", "foo/bar"},
		Success:    true,
	})

	testCases = append(testCases, TestCase{
		Name:      "screw.Mkdir/nonexistent",
		Argument:  "apricot",
		Operation: OpMkdir(screw.Mkdir),
		Success:   true,
	})

	testCases = append(testCases, TestCase{
		Name:       "screw.Mkdir/wrongcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpMkdir(screw.Mkdir),
		Error:      ErrorIs(screw.ErrCaseConflict),
	})

	testCases = append(testCases, TestCase{
		Name:       "screw.Mkdir/rightcase",
		DirsBefore: []string{"apricot"},
		Argument:   "apricot",
		Operation:  OpMkdir(screw.Mkdir),
		Error:      os.IsExist,
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.Mkdir/nonexistentparent",
		Argument:    "foo/bar",
		Operation:   OpMkdir(screw.Mkdir),
		Error:       os.IsNotExist,
		AbsentAfter: []string{"foo", "foo/bar"},
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.Mkdir/wrongcaseparent",
		DirsBefore:  []string{"FOO"},
		Argument:    "foo/bar",
		Operation:   OpMkdir(screw.Mkdir),
		Error:       os.IsNotExist,
		AbsentAfter: []string{"foo", "foo/bar"},

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "screw.Mkdir/rightcaseparent",
		DirsBefore: []string{"foo"},
		Argument:   "foo/bar",
		Operation:  OpMkdir(screw.Mkdir),
		Success:    true,
		DirsAfter:  []string{"foo", "foo/bar"},
	})

	//==========================
	// MkdirAll
	//==========================

	testCases = append(testCases, TestCase{
		Name:      "os.MkdirAll/nonexistent",
		Argument:  "apricot",
		Operation: OpMkdir(os.MkdirAll),
		Success:   true,
	})

	testCases = append(testCases, TestCase{
		Name:       "os.MkdirAll/mixedcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpMkdir(os.MkdirAll),
		Success:    true,
		DirsAfter:  []string{"APRICOT"},

		FSKind: FSCaseInsensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "os.MkdirAll/wrongcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpMkdir(os.MkdirAll),
		Success:    true,
		DirsAfter:  []string{"APRICOT", "apricot"},

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "os.MkdirAll/rightcase",
		DirsBefore: []string{"apricot"},
		Argument:   "apricot",
		Operation:  OpMkdir(os.MkdirAll),
		Success:    true,
		DirsAfter:  []string{"apricot"},
	})

	testCases = append(testCases, TestCase{
		Name:      "os.MkdirAll/nonexistentparent",
		Argument:  "foo/bar",
		Operation: OpMkdir(os.MkdirAll),
		Success:   true,
		DirsAfter: []string{"foo", "foo/bar"},
	})

	testCases = append(testCases, TestCase{
		Name:       "os.MkdirAll/mixedcaseparent",
		DirsBefore: []string{"FOO"},
		Argument:   "foo/bar",
		Operation:  OpMkdir(os.MkdirAll),
		Success:    true,
		DirsAfter:  []string{"FOO", "foo/bar"},

		FSKind: FSCaseInsensitive,
	})

	testCases = append(testCases, TestCase{
		Name:        "os.MkdirAll/wrongcaseparent",
		DirsBefore:  []string{"FOO"},
		Argument:    "foo/bar",
		Operation:   OpMkdir(os.MkdirAll),
		Error:       os.IsNotExist,
		AbsentAfter: []string{"FOO/bar"},

		FSKind: FSCaseSensitive,
	})

	testCases = append(testCases, TestCase{
		Name:       "os.MkdirAll/rightcaseparent",
		DirsBefore: []string{"foo"},
		Argument:   "foo/bar",
		Operation:  OpMkdir(os.MkdirAll),
		Success:    true,
		DirsAfter:  []string{"foo", "foo/bar"},
	})

	testCases = append(testCases, TestCase{
		Name:      "screw.MkdirAll/nonexistent",
		Argument:  "apricot",
		Operation: OpMkdir(screw.MkdirAll),
		Success:   true,
	})

	testCases = append(testCases, TestCase{
		Name:       "screw.MkdirAll/wrongcase",
		DirsBefore: []string{"APRICOT"},
		Argument:   "apricot",
		Operation:  OpMkdir(screw.MkdirAll),
		Error:      ErrorIs(screw.ErrCaseConflict),
		DirsAfter:  []string{"APRICOT"},
	})

	testCases = append(testCases, TestCase{
		Name:       "screw.MkdirAll/rightcase",
		DirsBefore: []string{"apricot"},
		Argument:   "apricot",
		Operation:  OpMkdir(screw.MkdirAll),
		Success:    true,
		DirsAfter:  []string{"apricot"},
	})

	testCases = append(testCases, TestCase{
		Name:      "screw.MkdirAll/nonexistentparent",
		Argument:  "foo/bar",
		Operation: OpMkdir(screw.MkdirAll),
		Success:   true,
		DirsAfter: []string{"foo", "foo/bar"},
	})

	testCases = append(testCases, TestCase{
		Name:        "screw.MkdirAll/wrongcaseparent",
		DirsBefore:  []string{"FOO"},
		Argument:    "foo/bar",
		Operation:   OpMkdir(screw.MkdirAll),
		Error:       ErrorIs(screw.ErrCaseConflict),
		AbsentAfter: []string{"FOO/bar"},
	})

	testCases = append(testCases, TestCase{
		Name:       "screw.MkdirAll/rightcaseparent",
		DirsBefore: []string{"foo"},
		Argument:   "foo/bar",
		Operation:  OpMkdir(screw.MkdirAll),
		Success:    true,
		DirsAfter:  []string{"foo", "foo/bar"},
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

		fullName := tc.Name + "/" + tc.FSKind.String()

		t.Run(fullName, func(t *testing.T) {
			assert := assert.New(t)

			dir, err := ioutil.TempDir("", "screw-tests")
			must(err)
			defer os.RemoveAll(dir)

			for _, name := range tc.DirsBefore {
				fullName := filepath.Join(dir, name)
				must(os.MkdirAll(fullName, 0o755))
			}

			for _, name := range tc.FilesBefore {
				fullName := filepath.Join(dir, name)
				must(os.MkdirAll(filepath.Dir(fullName), 0o755))

				f, err := os.Create(fullName)
				must(err)
				must(f.Close())
			}

			success, error := tc.Operation(filepath.Join(dir, tc.Argument))

			if tc.Success {
				assert.True(success, "operation should succeed")
				assert.NoError(error, "operation should not have an error")
			}

			if tc.Error != nil {
				assert.False(success, "operation should not succeed")
				assert.NotNil(error)
				if error != nil {
					assert.True(tc.Error(error), "error must pass test function")
				}
			}

			for _, ea := range tc.FilesAfter {
				stats, err := os.Stat(filepath.Join(dir, ea))
				assert.NoError(err, "%s must exist after", ea)
				assert.True(stats.Mode().IsRegular(), "%s must be a regular file after", ea)
			}

			for _, ea := range tc.DirsAfter {
				stats, err := os.Stat(filepath.Join(dir, ea))
				assert.NoError(err, "%s must exist after", ea)
				assert.True(stats.IsDir(), "%s must be a directory after", ea)
			}

			for _, ea := range tc.AbsentAfter {
				_, err := os.Stat(filepath.Join(dir, ea))
				assert.True(err != nil, "%s must be absent after", ea)
			}
		})
	}
}

func Test_IsActualCasing(t *testing.T) {
	if runtime.GOOS == "linux" {
		t.Skip()
	}

	assert := assert.New(t)

	tmpDir, err := ioutil.TempDir("", "screw-test-actual")
	must(err)

	reference := filepath.Join(tmpDir, "foo", "bar", "baz")

	err = os.MkdirAll(reference, 0o755)
	must(err)

	var is bool

	is, err = screw.IsActualCasing(reference)
	assert.NoError(err)
	assert.True(is)

	is, err = screw.IsActualCasing(filepath.Join(tmpDir, "foo", "bar", "BAZ"))
	assert.NoError(err)
	assert.False(is)

	is, err = screw.IsActualCasing(filepath.Join(tmpDir, "foo", "BAR", "baz"))
	assert.NoError(err)
	assert.False(is)

	is, err = screw.IsActualCasing(filepath.Join(tmpDir, "foo", "bar", "woops"))
	assert.Error(err)
	assert.False(is)

	var actual string

	actual, err = screw.GetActualCasing(reference)
	assert.NoError(err)
	assert.Equal(reference, actual)

	actual, err = screw.GetActualCasing(strings.ToUpper(reference))
	assert.NoError(err)
	assert.Equal(reference, actual)

	actual, err = screw.GetActualCasing(strings.ToLower(reference))
	assert.NoError(err)
	assert.Equal(reference, actual)

	actual, err = screw.GetActualCasing(filepath.Join(tmpDir, "FOO", "bar", "baz"))
	assert.NoError(err)
	assert.Equal(reference, actual)

	actual, err = screw.GetActualCasing(filepath.Join(tmpDir, "foo", "BAR", "baz"))
	assert.NoError(err)
	assert.Equal(reference, actual)
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
