# screw

[![Build Status](https://travis-ci.org/itchio/screw.svg?branch=master)](https://travis-ci.org/itchio/screw)
[![Build status](https://ci.appveyor.com/api/projects/status/8ewv7fb7myyb14r9?svg=true)](https://ci.appveyor.com/project/fasterthanlime/screw)

Screw case-insensitive filesystems.

Screw antivirus software locking files at random.

Let's abstract over them.

## Semantics

This table assumes `"apricot"` is passed to all those functions:

| Existing file name    | Operation          | `os` package           | `screw` package
|-----------------------|--------------------|------------------------|-----------------------
| (none)                | {S,Ls}tat          | ❌ os.ErrNotExist      | ❌ os.ErrNotExist
| "APRICOT"             | {S,Ls}tat          | ✅ stat "apricot"      | ❌ screw.ErrWrongCase
| "apricot"             | {S,Ls}tat          | ✅ stat "apricot"      | ✅ stat "apricot"
| (none)                | Open               | ❌ os.ErrNotExist      | ❌ os.ErrNotExist
| "APRICOT"             | Open               | ✅ open "apricot"      | ❌ screw.ErrWrongCase
| "apricot"             | Open               | ✅ open "apricot"      | ✅ open "apricot"
| (none)                | Create             | ✅ create "apricot"    | ✅ create "apricot"
| "APRICOT"             | Create             | ✅ truncate "APRICOT"  | ❌ screw.ErrWrongCase
| "apricot"             | Create             | ✅ truncate "apricot"  | ✅ truncate "apricot"

Note that the behavior is actually implemented in `OpenFile` so these are equivalent:

| Shorthand             | Actual call                                  |
|-----------------------|----------------------------------------------|
| Open(name)            | OpenFile(name, O_RDONLY, 0)                  |
| Create(name)          | OpenFile(name, O_RDWR|O_CREATE|O_TRUNC, 0666 |

Additional functions:

| Existing file name    | Operation             | Result
|-----------------------|-----------------------|--------------------
| (none)                | IsActualCase          | ❌ `os.ErrNotExist`
| "APRICOT"             | IsActualCase          | ✅ `false`
| "apricot"             | IsActualCase          | ✅ `true`

Additionally, `screw.Rename` contains retry logic on Windows (to sidestep spurious AV file locking),
and logic for older versions of Windows that don't support case-only renames.
