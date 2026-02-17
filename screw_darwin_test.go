//go:build darwin

package screw

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func stubRename(t *testing.T, fn func(oldpath, newpath string) error) {
	t.Helper()

	previous := osRename
	osRename = fn
	t.Cleanup(func() {
		osRename = previous
	})
}

func TestDoRename_DarwinSecondStageFailure(t *testing.T) {
	oldpath := "/tmp/old"
	newpath := "/tmp/new"
	secondStageErr := errors.New("second stage failed")

	var calls int
	var tmppath string

	stubRename(t, func(old, new string) error {
		calls++
		switch calls {
		case 1:
			if old != oldpath || new != newpath {
				t.Fatalf("unexpected first rename call: %q => %q", old, new)
			}
			return os.ErrExist
		case 2:
			if old != oldpath {
				t.Fatalf("unexpected second rename source: %q", old)
			}
			if !strings.HasPrefix(new, oldpath+"__rename_pid") {
				t.Fatalf("unexpected second rename target: %q", new)
			}
			tmppath = new
			return nil
		case 3:
			if old != tmppath || new != newpath {
				t.Fatalf("unexpected third rename call: %q => %q", old, new)
			}
			return secondStageErr
		case 4:
			if old != tmppath || new != oldpath {
				t.Fatalf("unexpected rollback call: %q => %q", old, new)
			}
			return nil
		default:
			t.Fatalf("unexpected extra rename call: %d", calls)
			return nil
		}
	})

	err := doRename(oldpath, newpath)
	if !errors.Is(err, secondStageErr) {
		t.Fatalf("expected second stage error, got: %v", err)
	}
	if calls != 4 {
		t.Fatalf("expected 4 rename calls, got: %d", calls)
	}
}

func TestDoRename_DarwinSecondStageAndRollbackFailure(t *testing.T) {
	oldpath := "/tmp/old"
	newpath := "/tmp/new"
	secondStageErr := errors.New("second stage failed")
	rollbackErr := errors.New("rollback failed")

	var calls int
	var tmppath string

	stubRename(t, func(old, new string) error {
		calls++
		switch calls {
		case 1:
			return os.ErrExist
		case 2:
			tmppath = new
			return nil
		case 3:
			if old != tmppath || new != newpath {
				t.Fatalf("unexpected third rename call: %q => %q", old, new)
			}
			return secondStageErr
		case 4:
			if old != tmppath || new != oldpath {
				t.Fatalf("unexpected rollback call: %q => %q", old, new)
			}
			return rollbackErr
		default:
			t.Fatalf("unexpected extra rename call: %d", calls)
			return nil
		}
	})

	err := doRename(oldpath, newpath)
	if !errors.Is(err, secondStageErr) {
		t.Fatalf("expected second stage error in chain, got: %v", err)
	}
	if !errors.Is(err, rollbackErr) {
		t.Fatalf("expected rollback error in chain, got: %v", err)
	}
	if calls != 4 {
		t.Fatalf("expected 4 rename calls, got: %d", calls)
	}
}
