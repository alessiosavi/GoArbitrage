package engine

import (
	"reflect"
	"testing"

	"github.com/alessiosavi/GoArbitrage/datastructure/market"
)

// initOrder is delegated to initalize a new map with the given key
func initOrder(keys []string) map[string][]market.MarketOrder {
	var commonOrder map[string][]market.MarketOrder = make(map[string][]market.MarketOrder)
	for _, key := range keys {
		commonOrder[key] = []market.MarketOrder{}
	}
	return commonOrder
}

var commonOrderKeys []string = []string{"adaeth", "btceth", "ltceth"}
var differentOrderKeys1 []string = []string{"AAAAA", "AAAAB", "AAAAC"}
var differentOrderKeys2 []string = []string{"AAAAD", "AAAAE", "AAAAF"}
var differentOrderKeys3 []string = []string{"AAAAG", "AAAAAH", "AAAAI"}
var differentOrderKeys4 []string = []string{"AAAAL", "AAAAM", "AAAAN"}

func Test_GetCommonCoinOK(t *testing.T) {
	var bitfinex market.Market = market.Market{MarketName: "BITFINEX", Asks: initOrder(commonOrderKeys)}
	var kraken market.Market = market.Market{MarketName: "KRAKEN", Asks: initOrder(commonOrderKeys)}
	var okcoin market.Market = market.Market{MarketName: "OKCOIN", Asks: initOrder(commonOrderKeys)}
	var gemini market.Market = market.Market{MarketName: "GEMINI", Asks: initOrder(commonOrderKeys)}
	commonPairs := GetCommonCoin(bitfinex, kraken, okcoin, gemini)
	if !reflect.DeepEqual(commonPairs, commonOrderKeys) {
		t.Fail()
	}
}

func Test_GetCommonCoinKO(t *testing.T) {
	var bitfinex market.Market = market.Market{MarketName: "BITFINEX", Asks: initOrder(differentOrderKeys1)}
	var kraken market.Market = market.Market{MarketName: "KRAKEN", Asks: initOrder(differentOrderKeys2)}
	var okcoin market.Market = market.Market{MarketName: "OKCOIN", Asks: initOrder(differentOrderKeys3)}
	var gemini market.Market = market.Market{MarketName: "GEMINI", Asks: initOrder(differentOrderKeys4)}
	commmonPairs := GetCommonCoin(bitfinex, kraken, okcoin, gemini)
	if len(commmonPairs) != 0 {
		t.Error("Pairs -> ", commmonPairs, " Len: ", len(commmonPairs))
	}
}
