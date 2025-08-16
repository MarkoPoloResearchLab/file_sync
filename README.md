# filez-sync

`filez-sync` is a standalone Go tool for **bidirectional file synchronization** with proper **3-way merges** for text files. It is designed to replace `unison` for cases where you need:

* Persistent merge history (per-file ancestor snapshots)
* Automatic use of `diff3` for conflict resolution
* Ignoring extra directories/files like `.git` or `node_modules`
* Optional backup of conflicting files before merging

Works for any pair of directories.

---

## Features

- **True 3-Way Merge** — if a common ancestor exists, uses GNU `diff3` for merging.
- **2-Way Merge Fallback** — if no ancestor exists, picks newer file by mtime, or embeds both with conflict markers.
- **Persistent Ancestor Store** — keeps ancestor blobs in a dedicated state directory for future merges.
- **Ignore Lists** — loaded from `.filezignore` to skip system trash folders, `.git`, `node_modules`, etc.
- **Optional Backups** — creates `.bak.a` / `.bak.b` before overwriting on conflicts.
- **Hash-Based Ancestor Tracking** — SHA-256 hashes ensure no accidental mix-ups.

---

## Installation

### Prerequisites
- Go 1.24+ (tested on Linux, macOS)
- (Optional) GNU `diff3` in `PATH` for better merge quality

```bash
go install ./cmd/filez-sync
```

This will produce a binary `filez-sync`.

### Docker

A prebuilt image is published to the GitHub Container Registry whenever Go files change on `master`.

```bash
docker pull ghcr.io/<OWNER>/filez-sync:latest
docker run --rm \
  -v /path/to/dir_a:/a \
  -v /path/to/dir_b:/b \
  -v /path/to/state:/state \
  ghcr.io/<OWNER>/filez-sync:latest \
  /a /b --state-dir /state --include '*'
```

Replace `<OWNER>` with the GitHub username or organization that owns the repository.

---

## Usage

```bash
filez-sync /path/to/dir_a /path/to/dir_b \
  --state-dir /path/to/state \
  --include '*'
```

By default, `filez-sync` walks both directories **recursively**. Use `--include` to match files by glob against each file's relative path and name. For example, `--include '*'` syncs everything, while `--include '*.txt'` limits the run to text files.

### Arguments

| Argument       | Required | Default | Description                                     |
| -------------- | -------- | ------- | ----------------------------------------------- |
| `root_a`       | ✅        | —       | First root directory                            |
| `root_b`       | ✅        | —       | Second root directory                           |
| `--state-dir`  | ✅        | —       | Directory for persistent sync state & ancestors |
| `--include`    | ❌        | —       | Glob for files to sync (defaults to an internal pattern) |
| `--no-backups` | ❌        | false   | Skip creation of `.bak.a` / `.bak.b`            |
| `--ignore-file` | ❌       | `.filezignore` | `.gitignore`-style file listing paths/patterns to ignore |

---

## How it Works

1. **First Run**

   * Every matching file in both roots is scanned.
   * Identical files → ancestor snapshot saved.
   * Different files without an ancestor → newer file wins (or conflict markers if mtimes close).
   * State is saved to `state.json` in `--state-dir`.

2. **Subsequent Runs**

   * For changed files, if an ancestor exists → run `diff3`.
   * If `diff3` fails/missing → fallback to simple merge with conflict markers.
   * Save merged result to both roots and update ancestor snapshot.

---

## Example

```bash
filez-sync \
  ~/Documents/ProjectA \
  /mnt/backup/ProjectA \
  --state-dir ~/.filez-sync-state \
  --include '*'
```

---

## Ignore Rules

`filez-sync` reads ignore rules from a `.gitignore`-style file on each run. The default
`.filezignore` includes:

```gitignore
# Directories
.git/
node_modules/
@eaDir/
#recycle/

# Files
.Trash*
.DS_Store
._*
Thumbs.db
desktop.ini
```

Provide a custom list by editing this file or passing `--ignore-file`:

```bash
filez-sync --ignore-file my_custom_ignore
```

---

## Exit Codes

* `0` — Sync completed successfully (some changes may have been made)
* `>0` — Fatal error

---

## Running Tests

The Go version includes an extensive test suite simulating:

* Two-way identical sync
* New file creation on each side
* Conflict resolution with and without ancestor
* 3-way merge correctness using `diff3`

Run tests with:

```bash
go test ./...
```

---

## License

MIT License — see `LICENSE` file.

