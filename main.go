package main

import (
	"log"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	constants "github.com/alessiosavi/GoArbitrage/datastructure/constants"
	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	"github.com/alessiosavi/GoArbitrage/engine"
	"github.com/alessiosavi/GoArbitrage/markets/bitfinex"
	"github.com/alessiosavi/GoArbitrage/markets/gemini"
	"github.com/alessiosavi/GoArbitrage/markets/kraken"
	"github.com/alessiosavi/GoArbitrage/markets/okcoin"
)

func main() {

	loggerMgr := initZapLog(zap.DebugLevel)
	zap.ReplaceGlobals(loggerMgr)
	defer loggerMgr.Sync() // flushes buffer, if any
	logger := loggerMgr.Sugar()
	logger.Infow("GoArbitrage started!")
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	initDataFolder()

	// Log configuration
	var bitfinex bitfinex.Bitfinex
	bitfinex.Init()
	bitfinex.GetPairsList()
	bitfinex.GetAllOrderBook()
	//log.Println(bitfinex.GetMarketData("etheur"))
	//log.Println(fmt.Sprintf("Bitfinex %#v\n", bitfinex))

	var okcoin okcoin.OkCoin
	okcoin.Init()
	okcoin.GetPairsList()
	okcoin.GetPairsDetails()
	okcoin.GetAllOrderBook()
	//log.Println(okcoin.GetMarketData("ETH-EUR"))
	// log.Println(fmt.Sprintf("OkCoin %#v\n", okcoin))

	var gemini gemini.Gemini
	gemini.Init()
	gemini.GetPairsList()
	gemini.GetAllOrderBook()
	gemini.GetPairsDetails()
	//log.Println(gemini.GetMarketData("bchbtc"))
	// log.Println(fmt.Sprintf("Gemini %#v\n", gemini))

	var kraken kraken.Kraken
	kraken.Init()
	kraken.GetPairsDetails()
	kraken.GetAllOrderBook()
	//log.Println(kraken.GetMarketData("ETHEUR"))
	// log.Println(fmt.Sprintf("Kraken %#v\n", kraken))

	zap.S().Warnf("Bitfinex: %f - %f", bitfinex.MakerFee, bitfinex.TakerFees)
	zap.S().Warnf("Gemini: %f - %f", gemini.MakerFee, gemini.TakerFees)
	zap.S().Warnf("OkCoin: %f - %f", okcoin.MakerFee, okcoin.TakerFees)
	var markets []market.Market

	// markets = append(markets, gemini.GetMarketsData())
	markets = append(markets, kraken.GetMarketsData())
	markets = append(markets, bitfinex.GetMarketsData())
	markets = append(markets, okcoin.GetMarketsData())

	pairs := engine.GetCommonCoin(markets...)
	zap.S().Infof("Common pairs: %v", pairs)
	for {
		for _, pair := range pairs {
			engine.Arbitrage(pair, markets)
		}
	}
}

func initZapLog(logLevel zapcore.Level) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.RFC3339NanoTimeEncoder
	logger, _ := config.Build()
	config.Level.SetLevel(logLevel)
	return logger
}

func initDataFolder() {
	if _, err := os.Stat(gemini.GEMINI_ORDERBOOK_DATA); os.IsNotExist(err) {
		zap.S().Debugw("Creating folder for GEMINI data ...")
		os.MkdirAll(constants.GEMINI_PATH, os.ModePerm)
		os.MkdirAll(gemini.GEMINI_ORDERBOOK_DATA, os.ModePerm)
	}

	if _, err := os.Stat(okcoin.OKCOIN_ORDERBOOK_DATA); os.IsNotExist(err) {
		zap.S().Debugw("Creating folder for OKCOIN data ...")
		os.MkdirAll(constants.OKCOIN_PATH, os.ModePerm)
		os.MkdirAll(okcoin.OKCOIN_ORDERBOOK_DATA, os.ModePerm)
	}

	if _, err := os.Stat(bitfinex.BITFINEX_ORDERBOOK_DATA); os.IsNotExist(err) {
		zap.S().Debugw("Creating folder for BITFINEX data ...")
		os.MkdirAll(constants.BITFINEX_PATH, os.ModePerm)
		os.MkdirAll(bitfinex.BITFINEX_ORDERBOOK_DATA, os.ModePerm)
	}

	if _, err := os.Stat(kraken.KRAKEN_ORDERBOOK_DATA); os.IsNotExist(err) {
		zap.S().Debugw("Creating folder for KRAKEN data ...")
		os.MkdirAll(constants.KRAKEN_PATH, os.ModePerm)
		os.MkdirAll(kraken.KRAKEN_ORDERBOOK_DATA, os.ModePerm)
	}
}
