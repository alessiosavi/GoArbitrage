package engine

import "github.com/alessiosavi/GoArbitrage/datastructure/market"

import "log"

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
