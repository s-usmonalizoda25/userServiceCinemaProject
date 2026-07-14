package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	*zap.Logger
}

func New() *Logger {
	config := zap.NewDevelopmentConfig()

	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	zapLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		panic("failed to initialize zap logger: " + err.Error())
	}
	return &Logger{Logger: zapLogger}
}
