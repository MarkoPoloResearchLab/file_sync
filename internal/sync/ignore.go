package sync

import (
	"path/filepath"

	"github.com/sabhiram/go-gitignore"
)

func shouldIgnore(rel string, isDir bool, ig *ignore.GitIgnore) bool {
	if ig == nil {
		return false
	}
	p := filepath.ToSlash(rel)
	if isDir {
		p += "/"
	}
	return ig.MatchesPath(p)
}
