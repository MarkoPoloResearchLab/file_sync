package main

import (
	"errors"
	"io/fs"
	"os"
	"strings"

	"github.com/MarkoPoloResearchLab/file_sync/internal/logging"
	syncpkg "github.com/MarkoPoloResearchLab/file_sync/internal/sync"
	"github.com/sabhiram/go-gitignore"
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
			ignoreFile := viper.GetString("ignore-file")

			if stateDir == "" {
				err := errors.New("--state-dir is required")
				logger.Error("missing state-dir", zap.Error(err))
				return err
			}

			ignoreMatcher, err := loadIgnoreMatcher(ignoreFile)
			if err != nil {
				logger.Error("read ignore file", zap.String("path", ignoreFile), zap.Error(err))
				return err
			}

			options := syncpkg.Options{
				RootAPath:                   args[0],
				RootBPath:                   args[1],
				StateDirectory:              stateDir,
				IncludeGlob:                 includePattern,
				CreateBackupsOnWrite:        !disableBackups,
				IgnoreMatcher:               ignoreMatcher,
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
	flags.String("include", "*.md", "glob for files to sync")
	flags.Bool("no-backups", false, "disable .bak files when overwriting")
	flags.String("log-level", "info", "log level")
	flags.String("ignore-file", ".filezignore", "path to .ignore-style file with ignore patterns")

	viper.SetEnvPrefix("FILEZ")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()

	viper.BindPFlag("state-dir", flags.Lookup("state-dir"))
	viper.BindPFlag("include", flags.Lookup("include"))
	viper.BindPFlag("no-backups", flags.Lookup("no-backups"))
	viper.BindPFlag("log-level", flags.Lookup("log-level"))
	viper.BindPFlag("ignore-file", flags.Lookup("ignore-file"))

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		var err error
		logger, err = logging.NewLogger()
		if err != nil {
			return err
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

func loadIgnoreMatcher(path string) (*ignore.GitIgnore, error) {
	ig, err := ignore.CompileIgnoreFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return ignore.CompileIgnoreLines(), nil
		}
		return nil, err
	}
	return ig, nil
}
