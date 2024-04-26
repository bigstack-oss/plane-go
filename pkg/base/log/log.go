package log

import (
	"fmt"
	"time"

	"github.com/bigstack-oss/plane-go/pkg/base/os"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var lvl int

type Logger struct {
	Log  *zap.Logger
	Logf *zap.SugaredLogger
}

func SetLogLevel(l int) {
	lvl = l
}

func GetLogLevel() zap.AtomicLevel {
	switch lvl {
	case 1:
		return zap.NewAtomicLevelAt(zap.ErrorLevel)
	case 2:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	case 3:
		return zap.NewAtomicLevelAt(zap.DebugLevel)
	default:
		return zap.NewAtomicLevelAt(zap.InfoLevel)
	}
}

func GetLogger(role string) *zap.Logger {
	cfg := zap.Config{
		Level:            GetLogLevel(),
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"role": role,
		},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "msg",
			TimeKey:     "ts",
			LevelKey:    "lvl",
			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02 15:04:05"))
			},
		},
	}

	logger, err := cfg.Build()
	if err != nil {
		fmt.Printf("error details of new logger: %s \n", err.Error())
		os.Exit(1)
	}

	return logger
}

func GetLoggers(role string) (*zap.Logger, *zap.SugaredLogger) {
	cfg := zap.Config{
		Level:            GetLogLevel(),
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields: map[string]interface{}{
			"role": role,
		},
		EncoderConfig: zapcore.EncoderConfig{
			MessageKey:  "msg",
			TimeKey:     "ts",
			LevelKey:    "lvl",
			EncodeLevel: zapcore.LowercaseLevelEncoder,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format("2006-01-02 15:04:05"))
			},
		},
	}

	logger, err := cfg.Build()
	if err != nil {
		fmt.Printf("error details of new logger: %s \n", err.Error())
		os.Exit(1)
	}

	return logger, logger.Sugar()
}
