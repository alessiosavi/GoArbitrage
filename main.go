package main

import (
	"fmt"

	"github.com/alessiosavi/GoArbitrage/markets/bitfinex"
)

func main() {

	// bitfinex.GetPairsList()
	var b bitfinex.Bitfinex
	b.GetPairsList()
	b.GetOrderBook()
	fmt.Printf("%+v\n", b)
}
