package engine

import (
	"log"

	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	"github.com/alessiosavi/GoArbitrage/markets/bitfinex"
	"github.com/alessiosavi/GoArbitrage/markets/gemini"
	"github.com/alessiosavi/GoArbitrage/markets/kraken"
	"github.com/alessiosavi/GoArbitrage/markets/okcoin"
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
	for i := range markets {

		switch markets[i].MarketName {
		case "KRAKEN":
			var kraken kraken.Kraken
			pair := kraken.ParsePair(pair)
			if kraken.GetOrderBook(pair) == nil {
				markets[i], err = kraken.GetMarketData(pair)
				if err != nil {
					log.Println("Unable to retrieve KRAKEN data ", err)
					return
				}
			}
		case "OKCOIN":
			var okcoin okcoin.OkCoin
			pair := okcoin.ParsePair(pair)
			if okcoin.GetOrderBook(pair) == nil {
				markets[i], err = okcoin.GetMarketData(pair)
				if err != nil {
					log.Println("Unable to retrieve OKCOIN data ", err)
					return
				}
			}
		case "BITFINEX":
			var bitfinex bitfinex.Bitfinex
			if bitfinex.GetOrderBook(pair) == nil {
				markets[i], err = bitfinex.GetMarketData(pair)
				if err != nil {
					log.Println("Unable to retrieve BITFINEX data ", err)
					return
				}
			}
		case "GEMINI":
			var gemini gemini.Gemini
			if gemini.GetOrderBook(pair) == nil {
				markets[i], err = gemini.GetMarketData(pair)
				if err != nil {
					log.Println("Unable to retrieve GEMINI data: ", err)
					return
				}
			}

		}
	}
	var minBuy market.Market = markets[0]
	var maxSell market.Market = markets[0]
	var pair1, pair2, pair3 string
	for i := 1; i < len(markets); i++ {
		pair1 = parsePair(pair, markets[i])
		pair2 = parsePair(pair, maxSell)
		pair3 = parsePair(pair, minBuy)
		log.Println("Checking markets [" + markets[i].MarketName + "] against [" + minBuy.MarketName + "] with pair: [" + pair1 + "] for BUY")
		//log.Println("Markets [", i, "] Asks:", markets[i].Asks[pair])
		if markets[i].Bids[pair1][0].Price < minBuy.Bids[pair3][0].Price && markets[i].MarketName != maxSell.MarketName {
			log.Println("Market [" + markets[i].MarketName + "] have a LESSER price than [" + minBuy.MarketName + "] FOR BUY")
			minBuy = markets[i]
		}
		log.Println("Checking markets [" + markets[i].MarketName + "] against [" + maxSell.MarketName + "] with pair: [" + pair2 + "] for BUY")
		if markets[i].Asks[pair1][0].Price > maxSell.Asks[pair2][0].Price && markets[i].MarketName != maxSell.MarketName {
			log.Println("Market [" + markets[i].MarketName + "] have a GREATER price than [" + maxSell.MarketName + "] FOR SELL")
			maxSell = markets[i]
		}
	}
	if minBuy.MarketName != maxSell.MarketName {
		pair2 = parsePair(pair, maxSell)
		pair3 = parsePair(pair, minBuy)
		log.Println("Arbitrage opportunity for pair: [" + pair + "]")
		log.Println("Buy Market:", minBuy.MarketName, "Price:", minBuy.Bids[pair3][0].Price, " Volume: ", minBuy.Bids[pair3][0].Volume)
		log.Println("Sell Market:", maxSell.MarketName, "Price:", maxSell.Asks[pair2][0].Price, " Volume: ", maxSell.Asks[pair2][0].Volume)
	}
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
