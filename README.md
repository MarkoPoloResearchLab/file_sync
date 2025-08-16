# filez-sync

`filez-sync` is a standalone Go tool for **bidirectional file synchronization** with proper **3-way merges** for Markdown (or any other text) files, designed to replace `unison` for cases where you need:

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
- **Ignore Lists** — loaded from `.filezignore` to skip system trash folders, `.obsidian`, `.git`, `node_modules`, etc.
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
  -v /path/to/vault_a:/a \
  -v /path/to/vault_b:/b \
  -v /path/to/state:/state \
  ghcr.io/<OWNER>/filez-sync:latest \
  /a /b --state-dir /state --include "*.md"
```

Replace `<OWNER>` with the GitHub username or organization that owns the repository.

---

## Usage

```bash
filez-sync /path/to/vault_a /path/to/vault_b \
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
  ~/Documents/ObsidianVault \
  /mnt/backup/ObsidianVault \
  --state-dir ~/.obsidian-sync-state
```

---

## Ignore Rules

`filez-sync` reads ignore rules from a `.gitignore`-style file on each run. The default
`.filezignore` includes:

```gitignore
# Directories
.obsidian/
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

