package main

import (
	"fmt"
	"os"
	"strings"

	syncmd "github.com/MarkoPoloResearchLab/file_sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "filez-sync [flags] <root_a> <root_b>",
	Short: "Synchronize files between two directories",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		stateDir := viper.GetString("state-dir")
		includePattern := viper.GetString("include")
		disableBackups := viper.GetBool("no-backups")

		if stateDir == "" {
			return fmt.Errorf("--state-dir is required")
		}

		options := syncmd.Options{
			RootAPath:            args[0],
			RootBPath:            args[1],
			StateDirectory:       stateDir,
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

		result, err := syncmd.RunSync(options)
		if err != nil {
			return err
		}

		fmt.Printf("changed=%d; actions=%v; diff3=%s\n",
			result.ChangedFileCount,
			result.ActionCounters,
			map[bool]string{true: "yes", false: "no"}[result.Diff3Available],
		)
		return nil
	},
}

func init() {
	flags := rootCmd.Flags()
	flags.String("state-dir", "", "directory for persistent state")
	flags.String("include", "*.md", "glob for files to sync")
	flags.Bool("no-backups", false, "disable .bak files when overwriting")

	viper.SetEnvPrefix("FILEZ")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.BindPFlag("state-dir", flags.Lookup("state-dir"))
	viper.BindPFlag("include", flags.Lookup("include"))
	viper.BindPFlag("no-backups", flags.Lookup("no-backups"))

	viper.SetConfigFile("config.yaml")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			fmt.Fprintf(os.Stderr, "error reading config file: %v\n", err)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
