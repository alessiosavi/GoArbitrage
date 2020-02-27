package utils

import (
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
