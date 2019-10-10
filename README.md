# screw

Screw case-insensitive filesystems.

Screw antivirus software locking files at random.

Let's abstract over them.

## Changes from 'os'

`screw` exposes a subset of `os`'s functions, with a few changes on Windows.
On other platforms, it makes no changes.

On Windows:

  - `screw.Rename` works properly for renames where *only the casing changes*.
  It appears that `os.Rename` also works for that scenario on Go 1.13 / Windows 10,
  but it wasn't always the case, so, for older versions, this will work (by first
  renaming to a temp name).
  - `screw.Rename` retries if it gets a permission denied error.
  - `screw.OpenFile` / `screw.Stat` / `screw.Lstat` return ErrNotFound
  for files that don't exist *with this exact casing*
  - `screw.Create` fails if another file exists with a different casing



