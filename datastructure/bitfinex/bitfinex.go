// This package will contains all the datastructure necessary for deal with Bitfinex market, internally
package bitfinex

type BitfinexPairsList struct {
	Pairs []string `json:"pairs"`
}

type BitfinexPairs struct {
	// Pair rappresent the two coins that are exchanged
	Pair string `json:"pair"`
	// MinOrder rappresent the minimum order allowed for the given pair
	MinOrder string `json:"minimum_order_size"`
	// MaxOrder rappresent the max order allowed for the given pair
	MaxOrder string `json:"maximum_order_size"`
	// PricePrecision rappresent the maximum number of significant digits for price in this pair
	PricePrecision int `json:"price_precision"`
}

type BitfinexOrder struct {
	Price  string `json:"price"`
	Amount string `json:"amount"`
}

type BitfinexOrderBook struct {
	Pair string          `json:"pair"`
	Bids []BitfinexOrder `json:"bids"`
	Asks []BitfinexOrder `json:"asks"`
}
