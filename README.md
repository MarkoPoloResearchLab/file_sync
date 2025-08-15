# File-sync

A standalone Go tool for **bidirectional file synchronization** with proper **3-way merges** for Markdown (or any other text) files, designed to replace `unison` for cases where you need:

* Persistent merge history (per-file ancestor snapshots)
* Automatic use of `diff3` for conflict resolution
* Ignoring extra directories/files like `.obsidian` or `node_modules`
* Optional backup of conflicting files before merging

Tested for syncing two Obsidian vaults across directories, but works for any two folders.

---

## Features

- **True 3-Way Merge** — if a common ancestor exists, uses GNU `diff3` for merging.
- **2-Way Merge Fallback** — if no ancestor exists, picks newer file by mtime, or embeds both with conflict markers.
- **Persistent Ancestor Store** — keeps ancestor blobs in a dedicated state directory for future merges.
- **Ignore Lists** — ignores system trash folders, `.obsidian`, `.git`, `node_modules`, etc.
- **Optional Backups** — creates `.bak.a` / `.bak.b` before overwriting on conflicts.
- **Hash-Based Ancestor Tracking** — SHA-256 hashes ensure no accidental mix-ups.

---

## Installation

### Prerequisites
- Go 1.22+ (tested on Linux, macOS)
- (Optional) GNU `diff3` in `PATH` for better merge quality

```bash
go install ./...
````

This will produce a binary `obsidian-3way-sync`.

---

## Usage

```bash
file-sync /path/to/vault_a /path/to/vault_b \
  --state-dir /path/to/state \
  --include "*.md"
```

### Arguments

| Argument       | Required | Default | Description                                     |
| -------------- | -------- | ------- | ----------------------------------------------- |
| `root_a`       | ✅        | —       | First root directory                            |
| `root_b`       | ✅        | —       | Second root directory                           |
| `--state-dir`  | ✅        | —       | Directory for persistent sync state & ancestors |
| `--include`    | ❌        | `*.md`  | Glob for files to sync                          |
| `--no-backups` | ❌        | false   | Skip creation of `.bak.a` / `.bak.b`            |

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
obsidian-3way-sync \
  ~/Documents/ObsidianVault \
  /mnt/backup/ObsidianVault \
  --state-dir ~/.obsidian-sync-state
```

---

## Ignore Rules

Directories ignored entirely:

```
.obsidian
.git
node_modules
@eaDir
#recycle
```

File name patterns ignored:

```
.Trash*
.DS_Store
._*
Thumbs.db
desktop.ini
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

