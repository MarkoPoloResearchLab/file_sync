package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestConfigurationLoadingAndLoggerInit(t *testing.T) {
	cases := []struct {
		name            string
		env             map[string]string
		cfg             string
		expectNoBackups bool
		debugEnabled    bool
	}{
		{
			name:            "EnvOverridesConfig",
                        env:             map[string]string{"ZYNC_NO_BACKUPS": "true"},
			cfg:             "state-dir: %s\nlog-level: debug\nno-backups: false\n",
			expectNoBackups: true,
			debugEnabled:    true,
		},
		{
			name:            "ConfigOnly",
			env:             map[string]string{},
			cfg:             "state-dir: %s\nlog-level: info\nno-backups: true\n",
			expectNoBackups: true,
			debugEnabled:    false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tmp := t.TempDir()
			cfg := filepath.Join(tmp, "config.yaml")
			stateDir := filepath.Join(tmp, "state")
			content := fmt.Sprintf(tc.cfg, stateDir)
			if err := os.WriteFile(cfg, []byte(content), 0o644); err != nil {
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

			for k, v := range tc.env {
				os.Setenv(k, v)
			}
			defer func() {
				for k := range tc.env {
					os.Unsetenv(k)
				}
			}()

			viper.Reset()
                        viper.SetEnvPrefix("ZYNC")
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
			if viper.GetBool("no-backups") != tc.expectNoBackups {
				t.Fatalf("no-backups=%v, want %v", viper.GetBool("no-backups"), tc.expectNoBackups)
			}
			if logger == nil || logger.Core().Enabled(zap.DebugLevel) != tc.debugEnabled {
				t.Fatalf("debug enabled=%v want %v", logger != nil && logger.Core().Enabled(zap.DebugLevel), tc.debugEnabled)
			}
		})
	}
}
