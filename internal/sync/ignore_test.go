package sync

import (
	"path/filepath"
	"testing"

	"github.com/sabhiram/go-gitignore"
)

func TestShouldIgnore(t *testing.T) {
	ig, err := ignore.CompileIgnoreFile(filepath.Join("..", "..", ".filezignore"))
	if err != nil {
		t.Fatalf("compile ignore file: %v", err)
	}

	cases := []struct {
		name   string
		path   string
		isDir  bool
		expect bool
	}{
		{"ObsidianDir", ".obsidian", true, true},
		{"ObsidianFile", ".obsidian/state.json", false, true},
		{"GitDir", ".git", true, true},
		{"NodeModulesDir", "node_modules", true, true},
		{"SynologyDir", "@eaDir", true, true},
		{"RecycleDir", "#recycle", true, true},
		{"TrashFile", ".Trash-123", false, true},
		{"DSStore", ".DS_Store", false, true},
		{"AppleDouble", "._foo", false, true},
		{"Thumbs", "Thumbs.db", false, true},
		{"DesktopIni", "desktop.ini", false, true},
		{"RegularFile", "notes/readme.md", false, false},
		{"RegularDir", "notes", true, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldIgnore(tc.path, tc.isDir, ig); got != tc.expect {
				t.Fatalf("shouldIgnore(%q, %v) = %v, want %v", tc.path, tc.isDir, got, tc.expect)
			}
		})
	}
}
