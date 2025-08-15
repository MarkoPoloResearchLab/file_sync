package logging

import (
	"testing"

	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func TestNewLoggerRespectsLogLevel(t *testing.T) {
	viper.Set("log-level", "debug")
	defer viper.Set("log-level", "")

	logger, err := NewLogger()
	if err != nil {
		t.Fatalf("NewLogger returned error: %v", err)
	}
	if !logger.Core().Enabled(zap.DebugLevel) {
		t.Fatalf("expected debug level to be enabled")
	}
}
