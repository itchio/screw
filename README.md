# screw

[![Build Status](https://travis-ci.org/itchio/screw.svg?branch=master)](https://travis-ci.org/itchio/screw)
[![Build status](https://ci.appveyor.com/api/projects/status/8ewv7fb7myyb14r9?svg=true)](https://ci.appveyor.com/project/fasterthanlime/screw)

**What do we have?** Case-preserving case-insensitive filesystems on Windows & macOS.

**What do we want?** Case-sensitive semantics, with one caveat: there cannot exist files that
differ in casing only.

## Semantics

First, note that core `screw` operations are implemented in `OpenFile`, so when we mean
`Open` or `Create`, we actually mean these:

| Shorthand             | Actual call                                     |
|-----------------------|-------------------------------------------------|
| Open(name)            | OpenFile(name, O_RDONLY, 0)                     |
| Create(name)          | OpenFile(name, O_RDWR\|O_CREATE\|O_TRUNC, 0666) |

The table assume `"apricot"` is passed to all those functions. The emoji meanings are as follows:

| Operation   | Existing file name    | `os` package           | `screw` package
|-------------|-----------------------|------------------------|-----------------------
| Stat, Lstat | (none)                | ❎ os.ErrNotExist      | (same)
|             | "apricot"             | ✅ stat "apricot"      | (same)
|             | "APRICOT"             | ⭕ stat "apricot"      | ❎ screw.ErrWrongCase
| Open        | (none)                | ❎ os.ErrNotExist      | (same)
|             | "apricot"             | ✅ open "apricot"      | (same)
|             | "APRICOT"             | ⭕ open "apricot"      | ❎ screw.ErrWrongCase
| Create      | (none)                | ✅ create "apricot"    | (same)
|             | "apricot"             | ✅ truncate "apricot"  | (same)
|             | "APRICOT"             | ⭕ truncate "APRICOT"  | ❎ screw.ErrWrongCase

Legend:

| Emoji | Meaning                  |
|-------|--------------------------|
| ✅   | Desirable return value    |
| ⭕   | Undesirable return value  |
| ❎   | Desirable error           |
| ❌   | Undesirable error         |

## Rename

On Windows, `screw.Rename` differs from `os.Rename` in two ways.

After a case-only rename (e.g. "apricot" => "APRICOT"), screw makes sure the file now has the expected casing.
If it doesn't, it attempts a two-step rename, ie.:

  * "apricot" => "apricot_rename_${pid}"
  * "apricot_rename_${pid}" => "APRICOT"
  
This seems unnecessary on recent versions of Windows 10 (as of October 2019), but it is the author's recollection
that this wasn't always the case.

Additionally, `screw.Rename` contains retry logic on Windows (to sidestep spurious AV file locking),
and logic for older versions of Windows that don't support case-only renames.

## API Additions

In addition to wrapping a lot of `os` functions, `screw` also provides these.

| Operation     | Existing file name    | Result
|---------------|-----------------------|--------------------
| IsActualCase  | (none)                | ❎ os.ErrNotExist
|               | "APRICOT"             | ✅ false
|               | "apricot"             | ✅ true
| GetActualCase | (none)                | ❎ os.ErrNotExist
|               | "APRICOT"             | ✅ "apricot" 
|               | "apricot"             | ✅ "apricot"

## Dependencies

On Windows, `screw` depends on `golang.org/x/sys/windows` to make the `FindFirstFile` syscall, instead of the legacy `syscall` package.

## Performance

Performance wasn't measured. `screw` is expected to be slower than straight up `os`, especially since each
`IsActualCase` / `GetActualCase` involves multiple calls to `FindFirstFile`
