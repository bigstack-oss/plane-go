package log

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestSetLogLevel(t *testing.T) {
	SetLogLevel(1)
	assert.Equal(t, 1, lvl, "failed to set log level")
}

func TestGetLogLevel(t *testing.T) {
	SetLogLevel(1)
	zapLogLevel := GetLogLevel()
	assert.Equal(t, zap.NewAtomicLevelAt(zap.ErrorLevel), zapLogLevel, "failed to set error log level")

	SetLogLevel(2)
	zapLogLevel = GetLogLevel()
	assert.Equal(t, zap.NewAtomicLevelAt(zap.InfoLevel), zapLogLevel, "failed to set info log level")

	SetLogLevel(3)
	zapLogLevel = GetLogLevel()
	assert.Equal(t, zap.NewAtomicLevelAt(zap.DebugLevel), zapLogLevel, "failed to set debug log level")

	SetLogLevel(9999)
	zapLogLevel = GetLogLevel()
	assert.Equal(t, zap.NewAtomicLevelAt(zap.InfoLevel), zapLogLevel, "failed to set default log level")
}

func TestGetLogger(t *testing.T) {
	SetLogLevel(2)
	testLogger := GetLogger("tester")
	testLogger.Info("test for logger")
	assert.Equal(t, true, testLogger.Core().Enabled(zap.InfoLevel), "failed to set log level")
}
