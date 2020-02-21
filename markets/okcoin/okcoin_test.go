package okcoin

import (
	"strings"
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
func Test_GetBookUrl(t *testing.T) {
	type testcase struct {
		pairs    string
		size     int
		depth    float64
		expected string
		number   int
	}

	cases := []testcase{
		testcase{pairs: "BTC-USDT", size: 10, depth: 0, expected: "https://www.okcoin.com/api/spot/v3/instruments/BTC-USDT/book?size=10", number: 1},
		testcase{pairs: "BTC-USDT", size: 10, depth: 0.001, expected: "https://www.okcoin.com/api/spot/v3/instruments/BTC-USDT/book?size=10&depth=0.001", number: 2},
	}

	for _, c := range cases {
		result := getBookURL(c.pairs, c.size, c.depth)
		if strings.Compare(result, c.expected) != 0 {
			t.Errorf("Received %v, expected %v [test n. %d]", result, c.expected, c.number)
		}
	}
}

func Test_GetPairsList(t *testing.T) {
	loggerMgr := initZapLog(zap.DebugLevel)
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // flushes buffer, if any
	var o OkCoin
	err := o.GetPairsList()
	if err != nil {
		t.Error(err)
	}
	t.Log(o.PairsName)
}

func Test_RetrieveTickers(t *testing.T) {
	loggerMgr := initZapLog(zap.DebugLevel)
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // flushes buffer, if any
	var o OkCoin
	if err := o.GetPairsList(); err != nil {
		t.Error(err)
	}
	t.Log(o.GetTickers())
}
