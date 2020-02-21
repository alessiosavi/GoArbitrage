package market

import "go.uber.org/zap"

type Market struct {
	// Name of the market
	MarketName string `json:"market_name"`
	// Asks and Bids contains the coin pair as a key and the value of the order
	Asks     map[string][]MarketOrder `json:"asks"`
	Bids     map[string][]MarketOrder `json:"bids"`
	MakerFee float64                  `json:"maker_fee"`
	TakerFee float64                  `json:"taker_fee"`
	Wallet   Wallet
}

type MarketOrder struct {
	Price     float64 `json:"price"`
	Volume    float64 `json:"volume"`
	MinVolume float64 `json:"min_volume"`
}

// MarketFee struct will save the type of the fee for every pairs.
// If IsPercent is true, than the fee will be calculated as a percent.
// In other case, we have to subtract the value cause is the pure fee
type MarketFee struct {
	IsPercent bool               `json:"is_percent"`
	Value     map[string]float64 `json:"value"`
}

// Wallet is delegated to save the coins related to the market
type Wallet struct {
	MarketName string             `json:"market_name"`
	Coins      map[string]float64 `json:"coins"`
}

// initDummyWalletCore is the core method for initialize a dummy wallet
func initDummyWalletCore(marketName string, currencies map[string]struct{}) Wallet {
	var w Wallet
	w.MarketName = marketName
	w.Coins = make(map[string]float64, len(currencies))
	for cName := range currencies {
		w.Coins[cName] = 999999
	}
	return w
}

// InitDummyWallet is delegated to initialize a dummy wallet for all the coin
func InitDummyWallet(markets []Market) {
	var c map[string]struct{} = make(map[string]struct{})
	for i := range markets {
		// Save the list of pair name
		for key := range markets[i].Asks {
			c[key] = struct{}{}
		}
		for key := range markets[i].Bids {
			c[key] = struct{}{}
		}
		markets[i].Wallet = initDummyWalletCore(markets[i].MarketName, c)
	}
}

// InitDummyWalletForPairs is delegated to initialize a dummy wallet with the given pairs
func InitDummyWalletForPairs(markets []Market, pairs []string) []Market {
	var c map[string]struct{} = make(map[string]struct{}, len(pairs))
	for i := range pairs {
		c[pairs[i]] = struct{}{}
	}

	for i := range markets {
		markets[i].Wallet = initDummyWalletCore(markets[i].MarketName, c)
		zap.S().Infof("Wallet: %+v", markets[i].Wallet)
	}

	return markets
}
