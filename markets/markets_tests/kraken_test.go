package markets_test

import (
	"testing"

	"github.com/alessiosavi/GoArbitrage/markets/kraken"
)

func Test_ParsePairKraken(t *testing.T) {

	type TestCase struct {
		Pair     string
		Expected string
		Number   int
	}

	var k kraken.Kraken
	cases := []TestCase{TestCase{Pair: "ethbtc", Expected: "ETHBTC", Number: 1}}

	for _, c := range cases {
		result := k.ParsePair(c.Pair)
		if result != c.Expected {
			t.Errorf("Received %v, expected %v [test n. %d]", result, c.Expected, c.Number)
		}

	}
}
