package syncmd

import (
	"path/filepath"
	"strings"
)

func shouldIgnorePath(relativePath string, ignorePrefixes []string) bool {
	norm := filepath.ToSlash(relativePath)
	for _, prefix := range ignorePrefixes {
		p := strings.TrimPrefix(filepath.ToSlash(prefix), "./")
		if norm == p || strings.HasPrefix(norm, p+"/") {
			return true
		}
	}
	return false
}

func shouldIgnoreName(fileName string, ignoreNames []string) bool {
	for _, pattern := range ignoreNames {
		matched, _ := filepath.Match(pattern, fileName)
		if matched {
			return true
		}
	}
	return false
}

