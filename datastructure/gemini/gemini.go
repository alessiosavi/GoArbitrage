package gemini

type GeminiPairsList struct {
	Pairs []string `json:"pairs"`
}

type GeminiOrder struct {
	Price     string `json:"price"`
	Volume    string `json:"volume"`
	Timestamp string `json:"timestamp"`
}

type GeminiOrderBook struct {
	Pair string        `json:"pair"`
	Bids []GeminiOrder `json:"bids"`
	Asks []GeminiOrder `json:"asks"`
}

type GeminiPairs struct {
	// Pair rappresent the two coins that are exchanged
	Pair string `json:"symbol"`
	// MinOrder rappresent the minimum order allowed for the given pair
	MinOrder          float32 `json:"min_order"`
	MinOrderIncrement float32 `json:"min_order_increment"`
	MinPriceIncrement float32 `json:"min_price_increment"`
}
