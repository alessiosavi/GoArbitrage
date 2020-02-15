package engine

import (
	"log"
	"sync"
	"time"

	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	"github.com/alessiosavi/GoArbitrage/markets/bitfinex"
	"github.com/alessiosavi/GoArbitrage/markets/gemini"
	"github.com/alessiosavi/GoArbitrage/markets/kraken"
	"github.com/alessiosavi/GoArbitrage/markets/okcoin"
	"go.uber.org/zap"
)

func GetCommonCoin(markets ...market.Market) []string {
	// commonPairs will save the list of pairs in common for the given markets
	var commonPairs []string

	// mapWithMaxLenght is the index of the given map that contains the higher number of pair,
	// that will be used to be compared against the other market
	var mapWithMaxLenght int = 0

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
			log.Println("Pair [" + key + "] Is in common in all market!")
			commonPairs = append(commonPairs, key)
		}
	}
	log.Println("Common pairs: ", commonPairs)
	return commonPairs
}

// Arbitrage is delegated to find the buy/sell
func Arbitrage(pair string, markets []market.Market) {
	var err error
	var wg sync.WaitGroup
	// Execute HTTP request in parallel
	start := time.Now()
	for i := range markets {
		switch markets[i].MarketName {
		case "KRAKEN":
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				var kraken kraken.Kraken
				pair := kraken.ParsePair(pair)
				if kraken.GetOrderBook(pair) == nil {
					markets[i], err = kraken.GetMarketData(pair)
					if err != nil {
						log.Println("Unable to retrieve KRAKEN data ", err)
						return
					}
				}
			}(i, &wg)
		case "OKCOIN":
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				var okcoin okcoin.OkCoin
				pair := okcoin.ParsePair(pair)
				if okcoin.GetOrderBook(pair) == nil {
					markets[i], err = okcoin.GetMarketData(pair)
					if err != nil {
						log.Println("Unable to retrieve OKCOIN data ", err)
						return
					}
				}
			}(i, &wg)
		case "BITFINEX":
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				var bitfinex bitfinex.Bitfinex
				if bitfinex.GetOrderBook(pair) == nil {
					markets[i], err = bitfinex.GetMarketData(pair)
					if err != nil {
						log.Println("Unable to retrieve BITFINEX data ", err)
						return
					}
				}
			}(i, &wg)
		case "GEMINI":
			wg.Add(1)
			go func(i int, wg *sync.WaitGroup) {
				defer wg.Done()
				var gemini gemini.Gemini
				if gemini.GetOrderBook(pair) == nil {
					markets[i], err = gemini.GetMarketData(pair)
					if err != nil {
						log.Println("Unable to retrieve GEMINI data: ", err)
						return
					}
				}
			}(i, &wg)
		}
	}
	wg.Wait()
	zap.S().Info("Time execution: ", time.Since(start))
	var minBuy market.Market = markets[0]
	var maxSell market.Market = markets[0]
	var pair1, pair2, pair3 string
	for i := 1; i < len(markets); i++ {
		pair1 = parsePair(pair, markets[i])
		pair2 = parsePair(pair, maxSell)
		pair3 = parsePair(pair, minBuy)
		zap.S().Debug("Checking markets [" + markets[i].MarketName + "] against [" + minBuy.MarketName + "] with pair: [" + pair1 + "] for BUY")
		if len(markets[i].Bids[pair1]) > 0 && len(minBuy.Bids[pair3]) > 0 && len(markets[i].Asks[pair1]) > 0 && len(maxSell.Asks[pair2]) > 0 {
			if markets[i].Bids[pair1][0].Price < minBuy.Bids[pair3][0].Price && markets[i].MarketName != maxSell.MarketName {
				zap.S().Debug("Market [" + markets[i].MarketName + "] have a LESSER price than [" + minBuy.MarketName + "] FOR BUY")
				minBuy = markets[i]
				minBuy.MakerFee = markets[i].MakerFee
				minBuy.TakerFee = markets[i].TakerFee
			}
			log.Println("Checking markets [" + markets[i].MarketName + "] against [" + maxSell.MarketName + "] with pair: [" + pair2 + "] for BUY")
			if markets[i].Asks[pair1][0].Price > maxSell.Asks[pair2][0].Price && markets[i].MarketName != maxSell.MarketName {
				zap.S().Debug("Market [" + markets[i].MarketName + "] have a GREATER price than [" + maxSell.MarketName + "] FOR SELL")
				maxSell = markets[i]
				maxSell.MakerFee = markets[i].MakerFee
				maxSell.TakerFee = markets[i].TakerFee
			}
			if minBuy.MarketName != maxSell.MarketName {
				pair2 = parsePair(pair, maxSell)
				pair3 = parsePair(pair, minBuy)
				if maxSell.Asks[pair2][0].Volume > maxSell.Asks[pair2][0].MinVolume && minBuy.Bids[pair3][0].Volume > minBuy.Bids[pair3][0].MinVolume {
					volume := getMin(maxSell.Asks[pair2][0].Volume, minBuy.Bids[pair3][0].Volume)
					buyTotal := volume * minBuy.Bids[pair3][0].Price
					buyTotal += percent(buyTotal, minBuy.TakerFee)
					sellTotal := volume * maxSell.Asks[pair2][0].Price
					sellTotal += percent(sellTotal, maxSell.TakerFee)
					zap.S().Infof("Arbitrage opportunity for pair [%s] with volume: %f\n", pair, volume)
					zap.S().Infof("Buy: %f Sell: %f | Difference: %f\n", buyTotal, sellTotal, sellTotal-buyTotal)
					zap.S().Infof("Buy Market: %s Price: %f Volume: %f\n", minBuy.MarketName, minBuy.Bids[pair3][0].Price, volume)
					zap.S().Infof("Sell Market: %s Price: %f Volume: %f\n", maxSell.MarketName, maxSell.Asks[pair2][0].Price, volume)
				}
			}
		}
	}
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
