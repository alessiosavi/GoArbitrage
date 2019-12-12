package bitfinex

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"path"
	"time"

	"github.com/alessiosavi/GoArbitrage/utils"

	constants "github.com/alessiosavi/GoArbitrage/datastructure"
	datastructure "github.com/alessiosavi/GoArbitrage/datastructure/bitfinex"
	fileutils "github.com/alessiosavi/GoGPUtils/files"
	req "github.com/alessiosavi/Requests"
)

const BITFINEX_PAIRS_URL string = `https://api.bitfinex.com/v1/symbols`
const BITFINEX_PAIRS_DETAILS_URL string = `https://api.bitfinex.com/v1/symbols_details`
const BITFINEX_ORDER_BOOK_URL string = `https://api.bitfinex.com/v1/book/`

var BITFINEX_PAIRS_DATA string = path.Join(constants.BITFINEX_PATH, "pairs_list.txt")
var BITFINEX_PAIRS_DETAILS string = path.Join(constants.BITFINEX_PATH, "pairs_info.txt")
var BITFINEX_ORDERBOOK_DATA string = path.Join(constants.BITFINEX_PATH, "orders/")

type Bitfinex struct {
	PairsNames []string
	Pairs      []datastructure.BitfinexPairs
	OrderBook  []datastructure.BitfinexOrderBook
}

// GetPairsList is delegated to retrieve the type of pairs in the Bitfinex market
func (b *Bitfinex) GetPairsList() []string {
	var request req.Request
	var pairs []string
	var data []byte
	var err error

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(BITFINEX_PAIRS_DATA) {
		log.Println("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(BITFINEX_PAIRS_DATA)
		if err != nil {
			log.Println("Error reading data: " + err.Error())
			return nil
		}
	} else {
		log.Println("Sendind request to [" + BITFINEX_PAIRS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(BITFINEX_PAIRS_URL, "GET", nil, false)
		if resp.Error != nil {
			log.Println("Error during http request. Err: " + resp.Error.Error())
			return nil
		}
		data = resp.Body
	}

	err = json.Unmarshal(data, &pairs)

	if err != nil {
		log.Println("Error during unmarshal! Err: " + err.Error())
	}

	b.PairsNames = pairs
	// Update the file with the new data
	utils.DumpStruct(pairs, BITFINEX_PAIRS_DATA)
	return pairs
}

// GetPairsDetails is delegated to retrieve the information related to all pairs for execute order
func (b *Bitfinex) GetPairsDetails() ([]datastructure.BitfinexPairs, error) {
	var request req.Request
	var pairsInfo []datastructure.BitfinexPairs
	var data []byte
	var err error

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(BITFINEX_PAIRS_DETAILS) {
		log.Println("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(BITFINEX_PAIRS_DETAILS)
		if err != nil {
			log.Println("Error reading data: " + err.Error())
			return nil, err
		}
	} else {
		log.Println("Sendind request to [" + BITFINEX_PAIRS_DETAILS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(BITFINEX_PAIRS_DETAILS_URL, "GET", nil, false)
		if resp.Error != nil {
			log.Println("Error during http request. Err: " + resp.Error.Error())
			return nil, resp.Error
		}
		data = resp.Body
	}

	err = json.Unmarshal(data, &pairsInfo)

	if err != nil {
		log.Println("Error during unmarshal! Err: " + err.Error())
	}

	b.Pairs = pairsInfo
	// Update the file with the new data
	utils.DumpStruct(pairsInfo, BITFINEX_PAIRS_DETAILS)
	return nil, nil
}

func (b *Bitfinex) GetOrderBook() {

	var request req.Request
	var orders []datastructure.BitfinexOrderBook
	var data []byte
	var err error

	for _, pair := range b.PairsNames {

		var orderbook datastructure.BitfinexOrderBook
		orderbook.Pair = pair
		file_data := path.Join(BITFINEX_ORDERBOOK_DATA, pair+".txt")
		// Avoid to call the HTTP api if the data are present
		if fileutils.FileExists(file_data) {
			log.Println("Data alredy present, avoiding to call the service")
			data, err = ioutil.ReadFile(file_data)
			if err != nil {
				log.Println("Error reading data: " + err.Error())
				return
			}

		} else {
			url := BITFINEX_ORDER_BOOK_URL + pair
			log.Println("Sendind request to [" + url + "]")
			// Call the HTTP method for retrieve the pairs
			resp := request.SendRequest(url, "GET", nil, false)
			if resp.Error != nil {
				log.Println("Error during http request. Err: " + resp.Error.Error())
				return
			}
			log.Println("Response -> " + resp.Dump())
			data = resp.Body
			time.Sleep(2 * time.Second)
		}

		err = json.Unmarshal(data, &orderbook)

		if err != nil {
			log.Println("Error during unmarshal! Err: " + err.Error())
		}
		orders = append(orders, orderbook)
		// Update the file with the new data
		utils.DumpStruct(orderbook, file_data)

	}

	b.OrderBook = orders

	// Update the file with the new data
	utils.DumpStruct(b.OrderBook, path.Join(constants.BITFINEX_PATH, "orders_all.txt"))
}
