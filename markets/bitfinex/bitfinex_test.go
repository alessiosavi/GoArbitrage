package bitfinex

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
func Test_GetTickers(t *testing.T) {
	loggerMgr := initZapLog(zap.DebugLevel)
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // flushes buffer, if any
	var err error
	var b Bitfinex
	if err = b.GetTickers(); err != nil {
		t.Error(err)
	}
	if len(b.Tickers) != 46 {
		t.Error("Different size ..")
	}
}
