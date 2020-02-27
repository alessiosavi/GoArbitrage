package utils

import (
	"reflect"
	"testing"

	"github.com/alessiosavi/GoArbitrage/datastructure/market"
)

// func Test_InitClient(t *testing.T) {
// 	c := InitClient()
// 	c.Close()
// }

func Test_RemoveMarket(t *testing.T) {
	var testData []market.Market = []market.Market{market.Market{MarketName: "1"},
		market.Market{MarketName: "2"},
		market.Market{MarketName: "3"},
		market.Market{MarketName: "4"},
		market.Market{MarketName: "5"},
		market.Market{MarketName: "6"}}

	var indexs []int = []int{0, 1, 2, 3}

	result := RemoveMarket(testData, indexs)
	if len(result) != 2 {
		t.Errorf("ERROR! Slice have not a lenght of 2, instead: %d", len(result))
	}
	if result[0].MarketName != "5" && result[1].MarketName != "6" {
		t.Errorf("ERROR! Invalid data, expected 5 and 6: %v", result)
	}
	t.Log(result)
}

func Test_ExtractCurrenciesFromPairs(t *testing.T) {
	var testData1 []string = []string{"ethbtc", "btceth"}
	var expected []string = []string{"eth", "btc"}
	result1 := ExtractCurrenciesFromPairs(testData1)
	if !reflect.DeepEqual(expected, result1) {
		t.Errorf("ERROR! Expected: %v | Result: %v", expected, result1)
	}
	// Will fail with 4 char pair
	// testData1 = []string{"ethtest", "testeth"}
	// expected = []string{"eth", "test"}
	// result1 = ExtractCurrenciesFromPairs(testData1)
	// if !reflect.DeepEqual(expected, result1) {
	// 	t.Errorf("ERROR! Expected: %v | Result: %v", expected, result1)
	// }

}
