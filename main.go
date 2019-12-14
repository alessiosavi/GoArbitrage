package main

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	constants "github.com/alessiosavi/GoArbitrage/datastructure"
	"github.com/alessiosavi/GoArbitrage/markets/bitfinex"
	"github.com/alessiosavi/GoArbitrage/markets/gemini"
	"github.com/alessiosavi/GoArbitrage/markets/kraken"
	"github.com/alessiosavi/GoArbitrage/markets/okcoin"
)

func main() {

	loggerMgr := initZapLog()
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // flushes buffer, if any
	logger := loggerMgr.Sugar()
	logger.Infow("GoArbitrage started!")
	initDataFolder()

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	// Log configuration

	// var b bitfinex.Bitfinex
	// b.SetFees()
	// b.GetPairsList()
	// b.GetOrderBook()
	// fmt.Printf("%+v\n", b)

	// var okcoin okcoin.OkCoin
	// okcoin.GetPairsList()
	// okcoin.GetPairsDetails()
	// okcoin.GetOrderBook()
	// log.Println(okcoin)

	// var gemini gemini.Gemini
	// gemini.GetPairsList()
	// gemini.GetOrderBook()
	// gemini.GetPairsDetails()
	// log.Println(gemini)

	var kraken kraken.Kraken

	kraken.GetPairsDetails()
	kraken.GetOrderBook()
	zap.S().Debug(kraken)
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

func initDataFolder() {
	if _, err := os.Stat(gemini.GEMINI_ORDERBOOK_DATA); os.IsNotExist(err) {
		zap.S().Debugw("Creating folder for GEMINI data ...")
		os.Mkdir(constants.GEMINI_PATH, os.ModePerm)
		os.Mkdir(gemini.GEMINI_ORDERBOOK_DATA, os.ModePerm)
	}

	if _, err := os.Stat(okcoin.OKCOIN_ORDERBOOK_DATA); os.IsNotExist(err) {
		zap.S().Debugw("Creating folder for GEMINI data ...")
		os.Mkdir(constants.OKCOIN_PATH, os.ModePerm)
		os.Mkdir(okcoin.OKCOIN_ORDERBOOK_DATA, os.ModePerm)
	}

	if _, err := os.Stat(bitfinex.BITFINEX_ORDERBOOK_DATA); os.IsNotExist(err) {
		zap.S().Debugw("Creating folder for GEMINI data ...")
		os.Mkdir(constants.BITFINEX_PATH, os.ModePerm)
		os.Mkdir(bitfinex.BITFINEX_ORDERBOOK_DATA, os.ModePerm)
	}

	if _, err := os.Stat(kraken.KRAKEN_ORDERBOOK_DATA); os.IsNotExist(err) {
		zap.S().Debugw("Creating folder for GEMINI data ...")
		os.Mkdir(constants.KRAKEN_PATH, os.ModePerm)
		os.Mkdir(kraken.KRAKEN_ORDERBOOK_DATA, os.ModePerm)
	}

}
