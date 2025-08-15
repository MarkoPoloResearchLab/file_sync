package logging

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewLogger creates a zap.Logger honoring the log-level configured via Viper.
func NewLogger() (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	levelStr := viper.GetString("log-level")
	if levelStr != "" {
		var lvl zapcore.Level
		if err := lvl.UnmarshalText([]byte(levelStr)); err != nil {
			return nil, err
		}
		cfg.Level = zap.NewAtomicLevelAt(lvl)
	}
	return cfg.Build()
}
