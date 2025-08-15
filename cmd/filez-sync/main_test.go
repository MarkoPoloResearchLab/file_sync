package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestConfigurationLoadingAndLoggerInit(t *testing.T) {
	tmp := t.TempDir()
	cfg := filepath.Join(tmp, "config.yaml")
	stateDir := filepath.Join(tmp, "state")
	if err := os.WriteFile(cfg, []byte("state-dir: "+stateDir+"\nlog-level: debug\n"), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	defer os.Chdir(cwd)
	if err := os.Chdir(tmp); err != nil {
		t.Fatalf("chdir: %v", err)
	}

	os.Setenv("FILEZ_NO_BACKUPS", "true")
	defer os.Unsetenv("FILEZ_NO_BACKUPS")

	viper.Reset()
	viper.SetEnvPrefix("FILEZ")
	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.AutomaticEnv()
	viper.BindPFlag("state-dir", rootCmd.Flags().Lookup("state-dir"))
	viper.BindPFlag("include", rootCmd.Flags().Lookup("include"))
	viper.BindPFlag("no-backups", rootCmd.Flags().Lookup("no-backups"))
	viper.BindPFlag("log-level", rootCmd.Flags().Lookup("log-level"))

	if err := rootCmd.PersistentPreRunE(rootCmd, []string{}); err != nil {
		t.Fatalf("pre-run: %v", err)
	}
	defer func() { logger = nil }()

	if got := viper.GetString("state-dir"); got != stateDir {
		t.Fatalf("state-dir not loaded: got %q want %q", got, stateDir)
	}
	if !viper.GetBool("no-backups") {
		t.Fatalf("expected no-backups from env")
	}
	if logger == nil || !logger.Core().Enabled(zap.DebugLevel) {
		t.Fatalf("logger not initialized with debug level")
	}
}
