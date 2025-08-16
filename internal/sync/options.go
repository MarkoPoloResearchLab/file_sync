package sync

import "github.com/sabhiram/go-gitignore"

// Options configures a synchronization run.
type Options struct {
	RootAPath                   string
	RootBPath                   string
	StateDirectory              string
	IncludeGlob                 string
	IgnoreMatcher               *ignore.GitIgnore
	CreateBackupsOnWrite        bool
	ConflictMtimeEpsilonSeconds float64
}
