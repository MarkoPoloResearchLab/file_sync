package main

import (
	"flag"
	"fmt"
	"os"

	"syncmd/internal/syncmd"
)

func main() {
	var stateDirectory string
	var includePattern string
	var disableBackups bool

	flag.StringVar(&stateDirectory, "state-dir", "", "directory for persistent state")
	flag.StringVar(&includePattern, "include", "*.md", "glob for files to sync")
	flag.BoolVar(&disableBackups, "no-backups", false, "disable .bak files when overwriting")
	flag.Parse()

	if flag.NArg() != 2 {
		fmt.Fprintln(os.Stderr, "usage: syncmd [--state-dir DIR] [--include GLOB] [--no-backups] <root_a> <root_b>")
		os.Exit(2)
	}

	rootA := flag.Arg(0)
	rootB := flag.Arg(1)

	if stateDirectory == "" {
		fmt.Fprintln(os.Stderr, "--state-dir is required")
		os.Exit(2)
	}

	options := syncmd.Options{
		RootAPath:            rootA,
		RootBPath:            rootB,
		StateDirectory:       stateDirectory,
		IncludeGlob:          includePattern,
		CreateBackupsOnWrite: !disableBackups,
		IgnorePathPrefixes: []string{
			".obsidian",
			".git",
			"node_modules",
			"@eaDir",
			"#recycle",
		},
		IgnoreFileNames: []string{
			".Trash*",
			".DS_Store",
			"._*",
			"Thumbs.db",
			"desktop.ini",
		},
		ConflictMtimeEpsilonSeconds: 1.0,
	}

	result, runErr := syncmd.RunSync(options)
	if runErr != nil {
			fmt.Fprintln(os.Stderr, runErr.Error())
			os.Exit(1)
	}

	fmt.Printf("changed=%d; actions=%v; diff3=%s\n",
		result.ChangedFileCount, result.ActionCounters, map[bool]string{true: "yes", false: "no"}[result.Diff3Available])
}

