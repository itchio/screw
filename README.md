# screw

[![Build Status](https://travis-ci.org/itchio/screw.svg?branch=master)](https://travis-ci.org/itchio/screw)
[![Build status](https://ci.appveyor.com/api/projects/status/8ewv7fb7myyb14r9?svg=true)](https://ci.appveyor.com/project/fasterthanlime/screw)

**What do we have?** case-preserving case-insensitive filesystems on Windows & macOS.

**What do we want?** case-*sensible* semantics on Windows, macOS & Linux.

Also, retry in case renaming/removing files fails because they're locked by an Antivirus.

## Intro

Throughout this documentation and code, we make the following assumptions:

  * that on Linux, we have a case-sensitive filesystem
  * that on Windows and macOS, we have a case-preserving, case-insensitive filesystem.

`screw` makes no effort to support [case-insensitive Linux
folders][ci-linux], or [case-sensitive Windows folders][cs-win], or [case-sensitive macOS partitions][cs-mac]

[ci-linux]: https://lwn.net/Articles/784041/
[cs-win]: https://devblogs.microsoft.com/commandline/per-directory-case-sensitivity-and-wsl/
[cs-mac]: https://apple.stackexchange.com/questions/8016/hfs-case-sensitive-or-case-insensitive-which-type-to-use-for-the-primary-dri

Now that we're clear on the scope of `screw`, let's agree on some common vocabulary.

## Case-sensitive filesystems (CS)

CS filesystems allow multiple files to exist, whose names differ only in casing.

For example, the following files may coexist:

```
- apricot
- Apricot
- APRICOT
```

They are three separate files, they can have separate contents, separate permissions, etc.
Deleting one of them doesn't touch the others.

Additionally, if only `apricot` exist, trying to access `APRICOT` will result in an error, since,
well, it doesn't exist.

Case-sensitive semantics are the easiest to remember.

## Case-preserving case-insensitive filesystems (CPCI)

The idea behind CPCI filesystems is that if a user creates file
named `apricot`, and later tries to access it as `APRICOT`, *we know what
they mean*, so we should just give them `apricot`.

The following properties follow:

  - Only *a single casing* of a name may exist at a time.
    - `apricot` and `APRICOT` may not both exist at the same time
  - Opening `APRICOT` opens `apricot`
  - Stat'ing `APRICOT` stats `apricot` *but it returns `APRICOT` as the name*
  - Deleting `APRICOT` deletes `apricot`
  - Creating `APRICOT` truncates `apricot`
  - Renaming `banana` to `APRICOT` will *replace* `apricot`
  - Renaming `apricot` to `APRICOT` is possible
    - But it may take extra steps, depending on the OS API

## Case-sensible semantics (CSBL)

`screw` tries to provide a *mostly case-sensitive* interface over CPCI filesystems.

If `APRICOT` exists, `screw` can pretend `apricot` doesn't - even though CPCI semantics
would say it does.

However, if `APRICOT` exists, creating `apricot` cannot be magically made to work.
Those two can't coexist under CPCI. In that case, `apricot` throws an error,
`screw.ErrCaseConflict`.

In short:

  * `screw` only lets you `Stat`, `Lstat`, `Open` or `Remove` a file **if you give its actual casing**
  * `screw` only lets you `Create` a file **if no casing variant exists**
  * `screw` has `RemoveAll` do nothing if **the exact casing you passed doesn't exist**

In the rest of the text, CPCI (case-preserving case-insensitive) semantics will be referred
to as "undesirable", and CSSB (case-sensible) semantics will be referred to as "desirable".

If you disagree, `screw` probably isn't for you!

## Changes from `os` package

First, note that core `screw` operations are implemented in `OpenFile`, so when we mean
`Open` or `Create`, we actually mean these:

| Shorthand             | Actual call                                      |
|-----------------------|--------------------------------------------------|
| Open(name)            | OpenFile(name, O_RDONLY, 0)                      |
| Create(name)          | OpenFile(name, O_RDWR\|O_CREATE\|O_TRUNC, 0o666) |

The tables in the rest of this document attribute meaning to some emojis:

| Emoji | Meaning                  |
|-------|--------------------------|
| ✅   | Desirable return value    |
| ⭕   | Undesirable return value  |
| ❎   | Desirable error           |
| ❌   | Undesirable error         |

The table assume `"apricot"` is passed to all those functions, and that the code
runs on a case-preserving, case-insensitive filesystem (Windows, macOS).

| Operation   | Existing file name    | `os` package (CPCI)    | `screw` package (CSBL)
|-------------|-----------------------|------------------------|-----------------------
| Stat, Lstat | (none)                | ❎ os.ErrNotExist      | 
|             | "apricot"             | ✅ stat "apricot"      | 
|             | "APRICOT"             | ⭕ stat "APRICOT"      | ❎ os.ErrNotExist
| Open        | (none)                | ❎ os.ErrNotExist      | 
|             | "apricot"             | ✅ open "apricot"      | 
|             | "APRICOT"             | ⭕ open "APRICOT"      | ❎ os.ErrNotExist
| Create      | (none)                | ✅ create "apricot"    | 
|             | "apricot"             | ✅ truncate "apricot"  | 
|             | "APRICOT"             | ⭕ truncate "APRICOT"  | ❎ screw.ErrCaseConflict

Destructive operations also behave differently:

| Operation   | Existing file name    | `os` package (CPCI)    | `screw` package (CSBL)
|-------------|-----------------------|------------------------|-----------------------
| Remove      | (none)                | ❎ os.ErrNotExist      | 
|             | "apricot"             | ✅ removes "apricot"   | 
|             | "APRICOT"             | ⭕ removes "APRICOT"   | ❎ os.ErrNotExist
| RemoveAll   | (none)                | ✅ does nothing        | 
|             | "apricot"             | ✅ removes "apricot"   | 
|             | "APRICOT"             | ⭕ removes "APRICOT"   | ✅ does nothing

The mkdir family behaves differently:

| Operation   | Existing dir name     | `os` package (CPCI)    | `screw` package (CSBL)
|-------------|-----------------------|------------------------|-----------------------
| Mkdir       | (none)                | ❎ mkdir "apricot"     | 
|             | "apricot/"            | ✅ os.ErrExist         | 
|             | "APRICOT/"            | ❌ os.ErrExist         | ❎ screw.ErrCaseConflict
| MkdirAll    | (none)                | ✅ mkdir "apricot"     | 
|             | "apricot/"            | ✅ does nothing        | 
|             | "APRICOT/"            | ⭕ does nothing        | ❎ screw.ErrCaseConflict

This also applies to subdirectories.

This table passes "foo/bar" to `Mkdir` and `MkdirAll`:

| Operation   | Existing dir name     | `os` package (CPCI)     | `screw` package (CSBL)
|-------------|-----------------------|-------------------------|-----------------------
| Mkdir       | (none)                | ❎ os.NotExist          | 
|             | "foo/"                | ✅ mkdir "foo/bar"      |
|             | "FOO/"                | ⭕ mkdir "FOO/bar"      | ❎ screw.ErrCaseConflict
| MkdirAll    | (none)                | ✅ mkdir -p "foo/bar"   | 
|             | "foo/"                | ✅ mkdir "foo/bar"      | 
|             | "FOO/"                | ⭕ mkdir "FOO/bar"      | ❎ screw.ErrCaseConflict

## Changes from `ioutil` package

`ioutil.ReadFile` and `ioutil.ReadDir` are included in `screw`

This table passes "apricot" to `ReadFile` and `ReadDir`

| Operation   | Existing file/dir name| `ioutil` package (CPCI) | `screw` package (CSBL)
|-------------|-----------------------|-------------------------|-----------------------
| ReadFile    | (none)                | ❎ os.NotExist          | 
|             | "apricot"             | ✅ read "apricot"       |
|             | "APRICOT"             | ⭕ read "APRICOT"       | ❎ os.NotExist
| ReadDir     | (none)                | ❎ os.NotExist          | 
|             | "apricot/"            | ✅ list "apricot/"      | 
|             | "APRICOT/"            | ⭕ list "APRICOT/"      | ❎ os.NotExist

## What about methods of *os.File ?

One of the undesired behaviors of CPCI is that the following code:

```go
func main() {
  // Note: the file `APRICOT` exists
  f, _ := os.Open("apricot")
  stats, _ := f.Stat()
  fmt.Println(stats.Name())
}
```

...will print `APRICOT`, not `apricot`.

What about `screw`? Shouldn't it have its own `File` type to prevent that?

In `screw`, that's not a problem, because the above code sample fails at the first line
with `os.ErrNotFound`, so any `os.FileInfo` obtained via screw contains the exact casing.

Similarly `file.Readdir()` and `file.Readdirname()` can only be called on a file
opened with its exact casing.

## API Additions

In addition to wrapping a lot of `os` functions, `screw` also provides these functions:

| Operation     | Existing file name    | Parameter                | Result
|---------------|-----------------------|--------------------------|--------------------------
| IsActualCase  | (none)                | "apricot"                | ❎ os.ErrNotExist
|               | "APRICOT"             | "apricot"                | ✅ false
|               | "apricot"             | "apricot"                | ✅ true
| GetActualCase | (none)                | "apricot"                | ❎ os.ErrNotExist
|               | "apricot"             | "apricot"                | ✅ "apricot" 
|               | "apricot"             | "APRICOT"                | ✅ "apricot"
|               | "apricot/seed"        | "apricot/SEED"           | ✅ "apricot/seed" 
|               | "apricot/seed"        | "APRICOT/seed"           | ✅ "apricot/seed"

**Important note**: contrary to the simplified table above, `GetActualCase` returns absolute paths,
not relative ones.

## Rename

On Windows, `screw.Rename` differs from `os.Rename` in two ways.

After a case-only rename (e.g. `apricot` => `APRICOT`), screw makes sure the file now has the expected casing.
If it doesn't, it attempts a two-step rename, ie.:

  * `apricot` => `apricot_rename_${pid}`
  * `apricot_rename_${pid}` => `APRICOT`
  
This seems unnecessary on recent versions of Windows 10 (as of October 2019), but it is the author's recollection
that this wasn't always the case.

Additionally, `screw.Rename` contains retry logic on Windows (to sidestep spurious AV file locking),
and logic for older versions of Windows that don't support case-only renames.

## UNC paths

UNC paths (like `\\?\C:\Windows\`, `\\SOMEHOST\\Share`) are untested and unsupported in `screw` at the time of this writing.

## Error wrapping

`screw` always wraps errors in a `*os.PathError`, to provide additional information as to which
operation caused the error, and on which file.

Comparing errors with `==` is a **bad idea**.

Using `errors.Is(e, screw.ErrCaseConflict)` works.

Using `os.IsNotExist(e)` also works with `screw`-returned errors.

## Dependencies

On Windows, `screw` depends on `golang.org/x/sys/windows` to make the `FindFirstFile` syscall, instead of the legacy `syscall` package.

## Performance

`screw` is expected to be slower than straight up `os`, especially since each
`IsActualCase` / `GetActualCase` involve multiple calls to `FindFirstFile` on Windows.

However, performance isn't a goal of `screw`, correctness is.
