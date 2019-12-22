package markets_test

import (
	"testing"

	"github.com/alessiosavi/GoArbitrage/markets/okcoin"
)

func Test_ParsePairOkCoin(t *testing.T) {

	type TestCase struct {
		Pair     string
		Expected string
		Number   int
	}

	var o okcoin.OkCoin
	cases := []TestCase{
		TestCase{Pair: "ethusd", Expected: "ETH-USD", Number: 1}, 
		TestCase{Pair: "adausd", Expected: "ADA-USD", Number: 2},
	 	TestCase{Pair: "btceurs", Expected: "BTC-EURS", Number: 3},
	  	TestCase{Pair: "btcusdt", Expected: "BTC-USDT", Number: 4},
	  	TestCase{Pair: "eurseur", Expected: "EURS-EUR", Number:5}}

	for _, c := range cases {
		result := o.ParsePair(c.Pair)
		if result != c.Expected {
			t.Errorf("Received %v, expected %v [test n. %d]", result, c.Expected, c.Number)
		}

	}
}
