package engine

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	"github.com/alessiosavi/GoArbitrage/markets/bitfinex"
	"github.com/alessiosavi/GoArbitrage/markets/gemini"
	"github.com/alessiosavi/GoArbitrage/markets/kraken"
	"github.com/alessiosavi/GoArbitrage/markets/okcoin"
	"github.com/alessiosavi/GoArbitrage/utils"
	"go.uber.org/zap"
)

type opportunity struct {
	MarketBuy     string          `json:"market_buy"`
	MarketSell    string          `json:"market_sell"`
	Pair          string          `json:"pair"`
	BuyPrice      float64         `json:"buy_price"`
	SellPrice     float64         `json:"sell_price"`
	Volume        float64         `json:"volume"`
	Earning       float64         `json:"earning"`
	Time          int64           `json:"time"`
	CurrentWallet []market.Wallet `json:"wallet"`
}

// GetCommonCoin : is delegated to retrieve the common pairs for the given markets
func GetCommonCoin(markets ...market.Market) []string {
	// commonPairs will save the list of pairs in common for the given markets
	var commonPairs []string

	// mapWithMaxLenght is the index of the given map that contains the higher number of pair,
	// that will be used to be compared against the other market
	var mapWithMaxLenght = 0

	// Retrieve the map with the max lenght
	for i := 1; i < len(markets); i++ {
		if len(markets[i].Asks) > len(markets[mapWithMaxLenght].Asks) {
			mapWithMaxLenght = i
		}
	}

	var isInCommon bool
	for key := range markets[mapWithMaxLenght].Asks {
		// Assume that is in common
		isInCommon = true
		for i := range markets {
			if isInCommon {
				// if the key is not present, than continue with the next key
				if _, found := markets[i].Asks[key]; !found {
					//log.Println("Pair [" + key + "] not found in market [" + markets[i].MarketName + "]")
					isInCommon = false
				} else {
					isInCommon = true
				}
			} else {
				break
			}
		}
		if isInCommon {
			zap.S().Debugf("Pair [%s] Is in common in all market!", key)
			commonPairs = append(commonPairs, key)
		}
	}
	zap.S().Infof("Common pairs: %v", commonPairs)
	return commonPairs
}

// Arbitrage is delegated to find the most relevant buy/sell opportunities for the given pair
func Arbitrage(pair string, markets *[]market.Market) {
	var err error
	var wg sync.WaitGroup
	// Execute HTTP request in parallel
	start := time.Now()
	var toRemove []int
	for i := range *markets {
		switch (*markets)[i].MarketName {
		case "KRAKEN":
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				var kraken kraken.Kraken
				var w market.Wallet = (*markets)[i].Wallet
				pair := kraken.ParsePair(pair)
				var makerFee, takerFee float64
				makerFee = (*markets)[i].MakerFee
				takerFee = (*markets)[i].TakerFee
				// FIXME: In case of error, the wallet will not be populated!
				if kraken.GetOrderBook(pair) == nil {
					(*markets)[i], err = kraken.GetMarketData(pair)
					if err == nil {
						(*markets)[i].Wallet = w
						(*markets)[i].MakerFee = makerFee
						(*markets)[i].TakerFee = takerFee
						return
					}
					zap.S().Warnf("Unable to retrieve KRAKEN data: %s", err.Error)
				}
				toRemove = append(toRemove, i)
			}(i, &wg)
		case "OKCOIN":
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				var okcoin okcoin.OkCoin
				var w market.Wallet = (*markets)[i].Wallet
				var makerFee, takerFee float64
				makerFee = (*markets)[i].MakerFee
				takerFee = (*markets)[i].TakerFee
				pair := okcoin.ParsePair(pair)
				if okcoin.GetOrderBook(pair) == nil {
					(*markets)[i], err = okcoin.GetMarketData(pair)
					if err == nil {
						(*markets)[i].Wallet = w
						(*markets)[i].MakerFee = makerFee
						(*markets)[i].TakerFee = takerFee
						return
					}
					zap.S().Warnf("Unable to retrieve OKCOIN data: %s", err.Error)
				}
				toRemove = append(toRemove, i)
			}(i, &wg)
		case "BITFINEX":
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				time.Sleep(time.Second * 2)
				defer wg.Done()
				var bitfinex bitfinex.Bitfinex
				var w market.Wallet = (*markets)[i].Wallet
				var makerFee, takerFee float64
				makerFee = (*markets)[i].MakerFee
				takerFee = (*markets)[i].TakerFee
				if bitfinex.GetOrderBook(pair) == nil {
					(*markets)[i], err = bitfinex.GetMarketData(pair)
					if err == nil {
						(*markets)[i].Wallet = w
						(*markets)[i].MakerFee = makerFee
						(*markets)[i].TakerFee = takerFee
						return
					}
					zap.S().Warnf("Unable to retrieve BITFINEX data: %s", err.Error)
				}
				toRemove = append(toRemove, i)
			}(i, &wg)
		case "GEMINI":
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				var gemini gemini.Gemini
				var w market.Wallet = (*markets)[i].Wallet
				var makerFee, takerFee float64
				makerFee = (*markets)[i].MakerFee
				takerFee = (*markets)[i].TakerFee
				if gemini.GetOrderBook(pair) == nil {
					(*markets)[i], err = gemini.GetMarketData(pair)
					if err == nil {
						(*markets)[i].Wallet = w
						(*markets)[i].MakerFee = makerFee
						(*markets)[i].TakerFee = takerFee
						return
					}
					zap.S().Warnf("Unable to retrieve GEMINI data: %s", err.Error)
				}
				toRemove = append(toRemove, i)
			}(i, &wg)
		}
	}
	wg.Wait()
	// Need to sort due to the concurrency
	sort.Ints(toRemove)
	*markets = utils.RemoveMarket(*markets, toRemove)
	zap.S().Info("Time execution: ", time.Since(start))

	if len(*markets) < 2 {
		zap.S().Warnf("Not have enough market to compare! Necessary at least 2, found: %d", len(*markets))
		return
	}
	var minBuy *market.Market = &(*markets)[0]
	var maxSell *market.Market = &(*markets)[0]

	var pair1, pair2, pair3 string
	var sb strings.Builder
	var opportunities []opportunity
	for i := 1; i < len(*markets); i++ {
		pair1 = parsePair(pair, (*markets)[i])
		pair2 = parsePair(pair, *maxSell)
		pair3 = parsePair(pair, *minBuy)
		_ = len((*markets)[i].Bids[pair1])
		zap.S().Debugf("Checking markets [%s] against [%s] with pair: [%s] for BUY", (*markets)[i].MarketName, minBuy.MarketName, pair1)
		if len((*markets)[i].Bids[pair1]) > 0 && len(minBuy.Bids[pair3]) > 0 && len(((*markets)[i].Asks[pair1])) > 0 && len(maxSell.Asks[pair2]) > 0 {
			if ((*markets)[i].Bids[pair1][0].Price) < minBuy.Bids[pair3][0].Price && ((*markets)[i].MarketName != maxSell.MarketName) {
				zap.S().Debugf("Market [%s] have a LESSER price than [%s] FOR BUY", (*markets)[i].MarketName, minBuy.MarketName)
				minBuy = &(*markets)[i]
				minBuy.MakerFee = (*markets)[i].MakerFee
				minBuy.TakerFee = (*markets)[i].TakerFee
			}
			zap.S().Debugf("Checking markets [%s] against [%s] with pair: [%s] for SELL", (*markets)[i].MarketName, maxSell.MarketName, pair2)
			if (*markets)[i].Asks[pair1][0].Price > maxSell.Asks[pair2][0].Price && (*markets)[i].MarketName != maxSell.MarketName {
				zap.S().Debugf("Market [%s] have a GREATER price than [%s] FOR SELL", (*markets)[i].MarketName, maxSell.MarketName)
				maxSell = &(*markets)[i]
				maxSell.MakerFee = (*markets)[i].MakerFee
				maxSell.TakerFee = (*markets)[i].TakerFee
			}
			if minBuy.MarketName != maxSell.MarketName {
				pair2 = parsePair(pair, *maxSell)
				pair3 = parsePair(pair, *minBuy)
				volume := getMin(maxSell.Asks[pair2][0].Volume, minBuy.Bids[pair3][0].Volume)
				buyTotal := volume * minBuy.Bids[pair3][0].Price
				buyTotal += percent(buyTotal, minBuy.TakerFee)
				sellTotal := volume * maxSell.Asks[pair2][0].Price
				sellTotal += percent(sellTotal, maxSell.TakerFee)
				if sellTotal-buyTotal > 0 {
					sb.WriteString(fmt.Sprintf("\nArbitrage opportunity for pair [%s] with volume: %f\n", pair, volume))
					sb.WriteString(fmt.Sprintf("Buy: %f Sell: %f | Difference: %f\n", buyTotal, sellTotal, sellTotal-buyTotal))
					sb.WriteString(fmt.Sprintf("Buy Market: %s Price: %f Volume: %f\n", minBuy.MarketName, minBuy.Bids[pair3][0].Price, volume))
					sb.WriteString(fmt.Sprintf("Sell Market: %s Price: %f Volume: %f\n", maxSell.MarketName, maxSell.Asks[pair2][0].Price, volume))
					zap.S().Info(sb.String())
					sb.Reset()
					var o opportunity
					o.BuyPrice = minBuy.Asks[pair3][0].Price
					o.SellPrice = maxSell.Bids[pair2][0].Price
					o.Pair = pair
					o.Volume = volume
					o.MarketBuy = minBuy.MarketName
					o.MarketSell = maxSell.MarketName
					o.Earning = sellTotal - buyTotal
					o.Time = time.Now().UnixNano()
					opportunities = append(opportunities, o)
				}
			}
		}
	}
	if len(opportunities) > 0 {
		var index = 0
		// Iterate the opportunities in order to extract only the most relevant
		for i := 1; i < len(opportunities); i++ {
			if opportunities[i].Earning > opportunities[index].Earning && opportunities[index].Earning > 0 {
				index = i
			}
		}

		zap.S().Infof("Before the reduce:  \nMinBuy: %+v \nMaxSell: %+v ", minBuy.Wallet, maxSell.Wallet)
		reduceWalletBalance(minBuy, maxSell, opportunities[index], pair)
		zap.S().Infof("After the reduce:  \nMinBuy: %+v \nMaxSell: %+v ", minBuy.Wallet, maxSell.Wallet)
		opportunities[index].CurrentWallet = getWalletFromMarkets(*markets)
		zap.S().Infof("Found the best opportunities for index %d: %+v", index, opportunities[index])
		zap.S().Debugf("All the data: %+v", opportunities)
		if data, err := json.Marshal(opportunities[index]); err == nil {
			if f, err := os.OpenFile("data.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
				defer f.Close()
				// writing the arbitrage opportunity
				if _, err := f.Write(data); err != nil {
					zap.S().Warnf("Unable to write to the data: %s", err.Error())
				} else {
					f.WriteString(",\n")
				}
			}
		}
	}
}

func dumpWallet(markets []market.Market) string {
	var w []byte
	w = append(w, []byte("[")...)
	for i := range markets {
		if data, err := json.Marshal(markets[i].Wallet); err == nil {
			w = append(w, data...)
			w = append(w, []byte(",")...)
		}
	}
	s := strings.TrimSuffix(string(w), ",")
	s += "]"
	return s
}

func getWalletFromMarkets(markets []market.Market) []market.Wallet {
	var wallets []market.Wallet = make([]market.Wallet, len(markets))
	for i := range markets {
		wallets[i] = markets[i].Wallet
	}
	return wallets
}

// reduceWalletBalance is delegated to remove the amount of coin from the sell market and increase the one related to the buy market
// If we are dealing with `ethusd` transaction, than we need to increase the `eth` and reduce the `usd`
func reduceWalletBalance(buy, sell *market.Market, operation opportunity, pair string) (market.Market, market.Market) {

	// a := "eth1btc"
	// b := a[len(a)-3:]// btc
	// c := a[:len(a)-len(b)] //eth1

	quoteCurrency := pair[len(pair)-3:]                 //b
	baseCurrency := pair[:len(pair)-len(quoteCurrency)] //c
	volume := operation.Volume
	buyPrice := operation.BuyPrice
	sellPrice := operation.SellPrice

	buy.Wallet.Coins[quoteCurrency] -= volume * buyPrice
	buy.Wallet.Coins[baseCurrency] += volume

	sell.Wallet.Coins[quoteCurrency] += volume * sellPrice
	sell.Wallet.Coins[baseCurrency] -= volume

	return *buy, *sell

}

func percent(percent float64, all float64) float64 {
	return (all * percent) / float64(100)
}

func getMin(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// parsePair is delegated to modify the standard lowercase pair into the related pair for the given market
func parsePair(pair string, market market.Market) string {
	switch market.MarketName {
	case "KRAKEN":
		var kraken kraken.Kraken
		pair = kraken.ParsePair(pair)
	case "OKCOIN":
		var okcoin okcoin.OkCoin
		pair = okcoin.ParsePair(pair)
	case "BITFINEX":
		// Nothing
	case "GEMINI":
		// Nothing
	}
	return pair
}
