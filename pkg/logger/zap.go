package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init() error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	var err error
	Log, err = config.Build()
	if err != nil {
		return fmt.Errorf("failed to initialize zap logger: %w", err)
	}

	Log.Info("zap logger initialized")
	return nil
}

func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
