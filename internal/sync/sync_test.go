package sync_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sabhiram/go-gitignore"
	syncpkg "github.com/MarkoPoloResearchLab/file_sync/internal/sync"
	"go.uber.org/zap"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	return string(data)
}

func defaultOptions(rootA string, rootB string, stateDir string) Options {
	ig := ignore.CompileIgnoreLines(
		".obsidian/",
		".git/",
		"node_modules/",
		"@eaDir/",
		"#recycle/",
		".Trash*",
		".DS_Store",
		"._*",
		"Thumbs.db",
		"desktop.ini",
	)
	return Options{
		RootAPath:                   rootA,
		RootBPath:                   rootB,
		StateDirectory:              state,
		IncludeGlob:                 "*.md",
		IgnoreMatcher:               ig,
		CreateBackupsOnWrite:        true,
		ConflictMtimeEpsilonSeconds: 1.0,
	}
}


func TestCreateFromSideA(t *testing.T) {
	rootA := newTempDir(t)
	rootB := newTempDir(t)
	stateDir := newTempDir(t)

	writeFile(t, filepath.Join(rootA, "Personal", "Note.md"), "hello")
	opts := defaultOptions(rootA, rootB, stateDir)

	res, err := RunSync(opts, zap.NewNop())
	if err != nil {
		t.Fatalf("sync err: %v", err)
	}
	if res.ChangedFileCount != 1 {
		t.Fatalf("changed count = %d", res.ChangedFileCount)
	}
	got := readFile(t, filepath.Join(rootB, "Personal", "Note.md"))
	if got != "hello" {
		t.Fatalf("unexpected content: %q", got)
	}
}

func TestEqualNoChange(t *testing.T) {
	rootA := newTempDir(t)
	rootB := newTempDir(t)
	stateDir := newTempDir(t)

	writeFile(t, filepath.Join(rootA, "a.md"), "same")
	writeFile(t, filepath.Join(rootB, "a.md"), "same")

	opts := defaultOptions(rootA, rootB, stateDir)
	res, err := RunSync(opts, zap.NewNop())
	if err != nil {
		t.Fatalf("sync err: %v", err)
	}
	if res.ChangedFileCount != 0 {
		t.Fatalf("expected no changes")
	}
}

func TestSeedConflictNewerWins(t *testing.T) {
	rootA := newTempDir(t)
	rootB := newTempDir(t)
	stateDir := newTempDir(t)

	writeFile(t, filepath.Join(rootA, "n.md"), "A1")
	writeFile(t, filepath.Join(rootB, "n.md"), "B1")

	// ensure distinct mtimes
	if runtime.GOOS != "windows" {
		os.Chtimes(filepath.Join(rootA, "n.md"), testTime(2000), testTime(2000))
		os.Chtimes(filepath.Join(rootB, "n.md"), testTime(3000), testTime(3000))
	}

	opts := defaultOptions(rootA, rootB, stateDir)
	res, err := RunSync(opts, zap.NewNop())
	if err != nil {
		t.Fatalf("sync err: %v", err)
	}
	if res.ActionCounters["merge(seed)"] != 1 {
		t.Fatalf("expected seed merge")
	}
	gotA := readFile(t, filepath.Join(rootA, "n.md"))
	gotB := readFile(t, filepath.Join(rootB, "n.md"))
	if gotA != gotB {
		t.Fatalf("sides differ after merge")
	}
	if gotB != "B1" && gotB != "<<<<<<<" {
		// newer wins unless times equal and markers emitted
		t.Fatalf("unexpected merged content: %q", gotB)
	}
}

func TestThreeWayAfterSeed(t *testing.T) {
	rootA := newTempDir(t)
	rootB := newTempDir(t)
	stateDir := newTempDir(t)

	writeFile(t, filepath.Join(rootA, "t.md"), "line1\n")
	writeFile(t, filepath.Join(rootB, "t.md"), "line1\n")

	opts := defaultOptions(rootA, rootB, stateDir)
	if _, err := RunSync(opts, zap.NewNop()); err != nil {
		t.Fatalf("initial sync: %v", err)
	}

	writeFile(t, filepath.Join(rootA, "t.md"), "line1\nA\n")
	writeFile(t, filepath.Join(rootB, "t.md"), "line1\nB\n")

	res, err := RunSync(opts, zap.NewNop())
	if err != nil {
		t.Fatalf("merge sync: %v", err)
	}
	if res.ChangedFileCount != 1 {
		t.Fatalf("expected one changed file, got %d", res.ChangedFileCount)
	}
	merged := readFile(t, filepath.Join(rootA, "t.md"))
	if !(strings.Contains(merged, "A") || strings.Contains(merged, "B")) {
		t.Fatalf("merged content not as expected: %q", merged)
	}
}

func TestIgnores(t *testing.T) {
	rootA := newTempDir(t)
	rootB := newTempDir(t)
	stateDir := newTempDir(t)

	writeFile(t, filepath.Join(rootA, ".obsidian", "state.json"), "{}")
	writeFile(t, filepath.Join(rootB, ".obsidian", "state.json"), "x")
	writeFile(t, filepath.Join(rootA, ".DS_Store"), "trash")
	writeFile(t, filepath.Join(rootA, "kept.md"), "K")
	opts := defaultOptions(rootA, rootB, stateDir)

	res, err := RunSync(opts, zap.NewNop())
	if err != nil {
		t.Fatalf("sync err: %v", err)
	}
	if res.ChangedFileCount != 1 {
		t.Fatalf("expected only one change for kept.md")
	}
	if _, err := os.Stat(filepath.Join(rootB, ".obsidian", "state.json")); err != nil {
		// ignored directory
	}
	if _, err := os.Stat(filepath.Join(rootB, ".DS_Store")); err == nil {
		t.Fatalf(".DS_Store should be ignored and not synced")
  }
}

func TestRunSync(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T, rootA, rootB, state string)
	}{
		{
			name: "CreateFromSideA",
			run: func(t *testing.T, rootA, rootB, state string) {
				writeFile(t, filepath.Join(rootA, "Personal", "Note.md"), "hello")
				opts := defaultOptions(rootA, rootB, state)
				res, err := syncpkg.RunSync(opts, zap.NewNop())
				if err != nil {
					t.Fatalf("sync err: %v", err)
				}
				if res.ChangedFileCount != 1 {
					t.Fatalf("changed count = %d", res.ChangedFileCount)
				}
				got := readFile(t, filepath.Join(rootB, "Personal", "Note.md"))
				if got != "hello" {
					t.Fatalf("unexpected content: %q", got)
				}
			},
		},
		{
			name: "EqualNoChange",
			run: func(t *testing.T, rootA, rootB, state string) {
				writeFile(t, filepath.Join(rootA, "a.md"), "same")
				writeFile(t, filepath.Join(rootB, "a.md"), "same")
				opts := defaultOptions(rootA, rootB, state)
				res, err := syncpkg.RunSync(opts, zap.NewNop())
				if err != nil {
					t.Fatalf("sync err: %v", err)
				}
				if res.ChangedFileCount != 0 {
					t.Fatalf("expected no changes")
				}
			},
		},
		{
			name: "SeedConflictNewerWins",
			run: func(t *testing.T, rootA, rootB, state string) {
				writeFile(t, filepath.Join(rootA, "n.md"), "A1")
				writeFile(t, filepath.Join(rootB, "n.md"), "B1")
				if runtime.GOOS != "windows" {
					os.Chtimes(filepath.Join(rootA, "n.md"), testTime(2000), testTime(2000))
					os.Chtimes(filepath.Join(rootB, "n.md"), testTime(3000), testTime(3000))
				}
				opts := defaultOptions(rootA, rootB, state)
				res, err := syncpkg.RunSync(opts, zap.NewNop())
				if err != nil {
					t.Fatalf("sync err: %v", err)
				}
				if res.ActionCounters["merge(seed)"] != 1 {
					t.Fatalf("expected seed merge")
				}
				gotA := readFile(t, filepath.Join(rootA, "n.md"))
				gotB := readFile(t, filepath.Join(rootB, "n.md"))
				if gotA != gotB {
					t.Fatalf("sides differ after merge")
				}
				if gotB != "B1" && !strings.Contains(gotB, "<<<<<<") {
					t.Fatalf("unexpected merged content: %q", gotB)
				}
			},
		},
		{
			name: "ThreeWayAfterSeed",
			run: func(t *testing.T, rootA, rootB, state string) {
				writeFile(t, filepath.Join(rootA, "t.md"), "line1\n")
				writeFile(t, filepath.Join(rootB, "t.md"), "line1\n")
				opts := defaultOptions(rootA, rootB, state)
				if _, err := syncpkg.RunSync(opts, zap.NewNop()); err != nil {
					t.Fatalf("initial sync: %v", err)
				}
				writeFile(t, filepath.Join(rootA, "t.md"), "line1\nA\n")
				writeFile(t, filepath.Join(rootB, "t.md"), "line1\nB\n")
				res, err := syncpkg.RunSync(opts, zap.NewNop())
				if err != nil {
					t.Fatalf("merge sync: %v", err)
				}
				if res.ChangedFileCount != 1 {
					t.Fatalf("expected one changed file, got %d", res.ChangedFileCount)
				}
				merged := readFile(t, filepath.Join(rootA, "t.md"))
				if !(strings.Contains(merged, "A") || strings.Contains(merged, "B")) {
					t.Fatalf("merged content not as expected: %q", merged)
				}
			},
		},
		{
			name: "Ignores",
			run: func(t *testing.T, rootA, rootB, state string) {
				writeFile(t, filepath.Join(rootA, ".obsidian", "state.json"), "{}")
				writeFile(t, filepath.Join(rootB, ".obsidian", "state.json"), "x")
				writeFile(t, filepath.Join(rootA, "kept.md"), "K")
				opts := defaultOptions(rootA, rootB, state)
				res, err := syncpkg.RunSync(opts, zap.NewNop())
				if err != nil {
					t.Fatalf("sync err: %v", err)
				}
				if res.ChangedFileCount != 1 {
					t.Fatalf("expected only one change for kept.md")
				}
				data, err := os.ReadFile(filepath.Join(rootB, ".obsidian", "state.json"))
				if err != nil {
					t.Fatalf("ignored file missing: %v", err)
				}
				if string(data) != "x" {
					t.Fatalf("ignored file was modified")
				}
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rootA := t.TempDir()
			rootB := t.TempDir()
			state := t.TempDir()
			tc.run(t, rootA, rootB, state)
		})
	}
}

func testTime(sec int64) time.Time {
	return time.Unix(sec, 0)
}
