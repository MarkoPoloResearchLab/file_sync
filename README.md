# zync

`zync` is a standalone Go tool for **bidirectional file synchronization** with proper **3-way merges** for Markdown (or any other text) files, designed to replace `unison` for cases where you need:

* Persistent merge history (per-file ancestor snapshots)
* Automatic use of `diff3` for conflict resolution
* Ignoring extra directories/files like `.obsidian` or `node_modules`
* Optional backup of conflicting files before merging

Designed for syncing any two directories without assumptions about their contents.

---

## Features

- **True 3-Way Merge** — if a common ancestor exists, uses GNU `diff3` for merging.
- **2-Way Merge Fallback** — if no ancestor exists, picks newer file by mtime, or embeds both with conflict markers.
- **Persistent Ancestor Store** — keeps ancestor blobs in a dedicated state directory for future merges.
- **Ignore Lists** — ignores system trash folders, `.obsidian`, `.git`, `node_modules`, etc.
- **Optional Backups** — creates `.bak.a` / `.bak.b` before overwriting on conflicts.
- **Hash-Based Ancestor Tracking** — SHA-256 hashes ensure no accidental mix-ups.

---

## Usage

You can obtain and run `zync` in several ways.

### Install with Go

Requires Go 1.24+ and optionally GNU `diff3` in your `PATH`.

```bash
go install github.com/MarkoPoloResearchLab/zync/cmd/zync@latest
```

The binary will be placed in `$(go env GOPATH)/bin` (or `GOBIN` if set).

### Download Prebuilt Binaries

Precompiled binaries are available on the [releases page](https://github.com/MarkoPoloResearchLab/zync/releases).
For example, on Linux x86_64:

```bash
curl -L https://github.com/MarkoPoloResearchLab/zync/releases/latest/download/zync_Linux_x86_64.tar.gz | tar -xz
sudo mv zync /usr/local/bin/
```

### Docker Image

A prebuilt image is published to the GitHub Container Registry whenever Go files change on `master`.

```bash
docker pull ghcr.io/<OWNER>/zync:latest
docker run --rm \
  -v /path/to/dir_a:/a \
  -v /path/to/dir_b:/b \
  -v /path/to/state:/state \
  ghcr.io/<OWNER>/zync:latest \
  /a /b --state-dir /state
```

Replace `<OWNER>` with the GitHub username or organization that owns the repository.

### CLI

```bash
zync /path/to/dir_a /path/to/dir_b \
  --state-dir /path/to/state
```

By default, `zync` walks both directories **recursively** and
synchronizes all files because `--include` defaults to `*`. The pattern
supplied to `--include` is matched against each file's relative path and file
name. Provide a more restrictive glob to limit which files are synchronized.
For example, to sync only Markdown files, use `--include "*.md"`.

```bash
# Only sync Markdown files
zync /path/to/dir_a /path/to/dir_b \
  --state-dir /path/to/state \
  --include "*.md"
```

### Arguments

| Argument       | Required | Default | Description                                     |
| -------------- | -------- | ------- | ----------------------------------------------- |
| `root_a`       | ✅        | —       | First root directory                            |
| `root_b`       | ✅        | —       | Second root directory                           |
| `--state-dir`  | ✅        | —       | Directory for persistent sync state & ancestors |
| `--include`    | ❌        | `*`     | Glob to restrict synced files                   |
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
zync \
  ~/Documents/ProjectA \
  /mnt/backup/ProjectA \
  --state-dir ~/.zync-state
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

