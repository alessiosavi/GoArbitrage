package engine

import (
	"log"

	"github.com/alessiosavi/GoArbitrage/datastructure/market"
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
					log.Println("Pair [" + key + "] not found in market [" + markets[i].MarketName + "]")
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

	var minBuy market.Market = markets[0]
	var maxSell market.Market = markets[0]
	for i := 1; i < len(markets); i++ {
		if markets[i].Bids[pair][0].Price > maxSell.Bids[pair][0].Price && markets[i].MarketName != maxSell.MarketName {
			log.Println("Market [" + markets[i].MarketName + "] have a GREATER price than [" + maxSell.MarketName + "] FOR BUY")
			maxSell = markets[i]
		}
		if markets[i].Asks[pair][0].Price < minBuy.Asks[pair][0].Price && markets[i].MarketName != maxSell.MarketName {
			log.Println("Market [" + markets[i].MarketName + "] have a LESSER price than [" + maxSell.MarketName + "] FOR SELL")
			minBuy = markets[i]
		}
	}
	if minBuy.MarketName != maxSell.MarketName {
		log.Println("Arbitrage opportunity for pair: [" + pair + "]")
		log.Println("Buy Market:", minBuy.MarketName, "Price:", minBuy.Asks[pair][0].Price, " Volume: ", minBuy.Asks[pair][0].Volume)
		log.Println("Sell Market:", maxSell.MarketName, "Price:", maxSell.Bids[pair][0].Price, " Volume: ", maxSell.Bids[pair][0].Volume)
	}
}
