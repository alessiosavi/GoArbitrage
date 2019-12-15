package okcoin

import "testing"

import "strings"

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
		result := getBookUrl(c.pairs, c.size, c.depth)
		if strings.Compare(result, c.expected) != 0 {
			t.Errorf("Received %v, expected %v [test n. %d]", result, c.expected, c.number)
		}
	}
}
