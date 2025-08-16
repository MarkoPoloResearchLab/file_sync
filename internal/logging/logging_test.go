package logging

import (
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestNewLoggerRespectsLogLevel(t *testing.T) {
	cases := []struct {
		name        string
		level       string
		enableDebug bool
		wantErr     bool
	}{
		{"debug", "debug", true, false},
		{"info", "info", false, false},
		{"invalid", "bogus", false, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Set("log-level", tc.level)
			defer viper.Set("log-level", "")

			logger, err := NewLogger()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for log level %q", tc.level)
				}
				return
			}
			if err != nil {
				t.Fatalf("NewLogger returned error: %v", err)
			}
			if logger.Core().Enabled(zap.DebugLevel) != tc.enableDebug {
				t.Fatalf("debug enabled = %v, want %v", logger.Core().Enabled(zap.DebugLevel), tc.enableDebug)
			}
		})
	}
}
