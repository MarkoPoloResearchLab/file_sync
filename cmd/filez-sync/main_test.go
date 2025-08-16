package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestEnvironmentLoadingAndLoggerInit(t *testing.T) {
	tmp := t.TempDir()
	stateDir := filepath.Join(tmp, "state")
	os.Setenv("FILEZ_STATE_DIR", stateDir)
	os.Setenv("FILEZ_NO_BACKUPS", "true")
	os.Setenv("FILEZ_LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("FILEZ_STATE_DIR")
		os.Unsetenv("FILEZ_NO_BACKUPS")
		os.Unsetenv("FILEZ_LOG_LEVEL")
	}()

	viper.Reset()
	viper.SetEnvPrefix("FILEZ")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	viper.BindPFlag("state-dir", rootCmd.Flags().Lookup("state-dir"))
	viper.BindPFlag("include", rootCmd.Flags().Lookup("include"))
	viper.BindPFlag("no-backups", rootCmd.Flags().Lookup("no-backups"))
	viper.BindPFlag("log-level", rootCmd.Flags().Lookup("log-level"))
	viper.BindPFlag("ignore-file", rootCmd.Flags().Lookup("ignore-file"))

	if err := rootCmd.PersistentPreRunE(rootCmd, []string{}); err != nil {
		t.Fatalf("pre-run: %v", err)
	}
	defer func() { logger = nil }()

	if got := viper.GetString("state-dir"); got != stateDir {
		t.Fatalf("state-dir not loaded: got %q want %q", got, stateDir)
	}
	if !viper.GetBool("no-backups") {
		t.Fatalf("no-backups flag not loaded")
	}
	if logger == nil || !logger.Core().Enabled(zap.DebugLevel) {
		t.Fatalf("debug logging not enabled")
	}
}
