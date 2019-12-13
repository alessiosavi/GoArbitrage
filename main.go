package main

import (
	"log"

	"github.com/alessiosavi/GoArbitrage/markets/okcoin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {

	loggerMgr := initZapLog()
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // flushes buffer, if any
	logger := loggerMgr.Sugar()
	logger.Infow("GoArbitrage started!")
	// Log configuration

	// var b bitfinex.Bitfinex
	// b.SetFees()
	// b.GetPairsList()
	// b.GetOrderBook()
	// fmt.Printf("%+v\n", b)

	var okcoin okcoin.OkCoin

	okcoin.GetPairsList()
	okcoin.GetPairsDetails()
	okcoin.GetOrderBook()

	log.Println(okcoin)

}

func initZapLog() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	//zapcore.ISO8601TimeEncoder
	logger, _ := config.Build()
	return logger
}
