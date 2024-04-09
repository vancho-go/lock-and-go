package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger SugaredLogger обертка.
type Logger = zap.SugaredLogger

// New генерирует логгер.
func New(logLevel string) (*Logger, error) {
	config := zap.NewProductionConfig()

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		return nil, err
	}
	config.Level = zap.NewAtomicLevelAt(level)

	config.Encoding = "console" // "console" для более читаемого вывода
	config.EncoderConfig = zap.NewProductionEncoderConfig()
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	logger, err := config.Build()
	if err != nil {
		return nil, err
	}

	return logger.Sugar(), nil
}
