package main

import (
	"errors"
	"os"
	"strings"

	"github.com/MarkoPoloResearchLab/file_sync/internal/logging"
	syncpkg "github.com/MarkoPoloResearchLab/file_sync/internal/sync"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	logger  *zap.Logger
	rootCmd = &cobra.Command{
		Use:   "filez-sync [flags] <root_a> <root_b>",
		Short: "Synchronize files between two directories",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			stateDir := viper.GetString("state-dir")
			includePattern := viper.GetString("include")
			disableBackups := viper.GetBool("no-backups")

			if stateDir == "" {
				err := errors.New("--state-dir is required")
				logger.Error("missing state-dir", zap.Error(err))
				return err
			}

			options := syncpkg.Options{
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

			result, err := syncpkg.RunSync(options, logger)
			if err != nil {
				logger.Error("synchronization failed", zap.Error(err))
				return err
			}

			logger.Info("synchronization completed",
				zap.Int("changed", result.ChangedFileCount),
				zap.Any("actions", result.ActionCounters),
				zap.Bool("diff3", result.Diff3Available),
			)

			return nil
		},
	}
)

func init() {
	flags := rootCmd.Flags()
	flags.String("state-dir", "", "directory for persistent state")
	flags.String("include", "*", "glob to restrict synced files (default '*')")
	flags.Bool("no-backups", false, "disable .bak files when overwriting")
	flags.String("log-level", "info", "log level")

	viper.SetEnvPrefix("FILEZ")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.BindPFlag("state-dir", flags.Lookup("state-dir"))
	viper.BindPFlag("include", flags.Lookup("include"))
	viper.BindPFlag("no-backups", flags.Lookup("no-backups"))
	viper.BindPFlag("log-level", flags.Lookup("log-level"))

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		viper.SetConfigFile("config.yaml")
		cfgErr := viper.ReadInConfig()

		var err error
		logger, err = logging.NewLogger()
		if err != nil {
			return err
		}

		if cfgErr != nil {
			if _, ok := cfgErr.(viper.ConfigFileNotFoundError); !ok {
				logger.Error("error reading config file", zap.Error(cfgErr))
			}
		}
		return nil
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		if logger != nil {
			logger.Error("command failed", zap.Error(err))
			_ = logger.Sync()
		} else {
			os.Stderr.WriteString(err.Error() + "\n")
		}
		os.Exit(1)
	}
	if logger != nil {
		_ = logger.Sync()
	}
}
