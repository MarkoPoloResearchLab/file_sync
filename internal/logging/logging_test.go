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
	}{
		{"debug", "debug", true},
		{"info", "info", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			viper.Set("log-level", tc.level)
			defer viper.Set("log-level", "")

			logger, err := NewLogger()
			if err != nil {
				t.Fatalf("NewLogger returned error: %v", err)
			}
			if logger.Core().Enabled(zap.DebugLevel) != tc.enableDebug {
				t.Fatalf("debug enabled = %v, want %v", logger.Core().Enabled(zap.DebugLevel), tc.enableDebug)
			}
		})
	}
}
