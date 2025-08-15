package syncmd

// Options configures a synchronization run.
type Options struct {
	RootAPath                    string
	RootBPath                    string
	StateDirectory               string
	IncludeGlob                  string
	IgnorePathPrefixes           []string
	IgnoreFileNames              []string
	CreateBackupsOnWrite         bool
	ConflictMtimeEpsilonSeconds  float64
}

