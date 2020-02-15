package market

type Market struct {
	// Name of the market
	MarketName string `json:"market_name"`
	// Asks and Bids contains the coin pair as a key and the value of the order
	Asks      map[string][]MarketOrder `json:"asks"`
	Bids      map[string][]MarketOrder `json:"bids"`
	MinVolume float64                  `json:"min_volume"`
	MakerFee  float64                  `json:"maker_fee"`
	TakerFee  float64                  `json:"taker_fee"`
}

type MarketOrder struct {
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}

// MarketFee struct will save the type of the fee for every pairs.
// If IsPercent is true, than the fee will be calculated as a percent.
// In other case, we have to subtract the value cause is the pure fee
type MarketFee struct {
	IsPercent bool               `json:"is_percent"`
	Value     map[string]float64 `json:"value"`
}
