package okcoin

import "time"

// OkCoinPair is used for retrievethe list of pairs that can be traded on OKCOIN
type OkCoinPair struct {
	ProductID string `json:"product_id"`
}

// OkCoinPairs contains the information related to the pairs
type OkCoinPairs struct {
	Pair          string `json:"pair"`
	BaseCurrency  string `json:"base_currency"`
	MinSize       string `json:"min_size"`
	QuoteCurrency string `json:"quote_currency"`
	SizeIncrement string `json:"size_increment"`
	TickSize      string `json:"tick_size"`
}

type OkCoinOrderBook struct {
	Pair      string     `json:"pair"`
	Asks      [][]string `json:"asks"`
	Bids      [][]string `json:"bids"`
	Timestamp time.Time  `json:"timestamp"`
}
