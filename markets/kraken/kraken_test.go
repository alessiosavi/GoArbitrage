package kraken

import (
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func initZapLog(logLevel zapcore.Level) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logger, _ := config.Build()
	config.Level.SetLevel(logLevel)
	return logger
}
func Test_RetrieveTickers(t *testing.T) {
	loggerMgr := initZapLog(zap.DebugLevel)
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // flushes buffer, if any
	var k Kraken
	if err := k.GetTickers(); err != nil {
		t.Error(err)
	}
}
