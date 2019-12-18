package bitfinex

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alessiosavi/GoArbitrage/utils"
	"go.uber.org/zap"

	datastructure "github.com/alessiosavi/GoArbitrage/datastructure/bitfinex"
	constants "github.com/alessiosavi/GoArbitrage/datastructure/constants"
	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	fileutils "github.com/alessiosavi/GoGPUtils/files"
	req "github.com/alessiosavi/Requests"
)

const BITFINEX_PAIRS_URL string = `https://api.bitfinex.com/v1/symbols`
const BITFINEX_PAIRS_DETAILS_URL string = `https://api.bitfinex.com/v1/symbols_details`
const BITFINEX_ORDER_BOOK_URL string = `https://api.bitfinex.com/v1/book/`

var BITFINEX_PAIRS_DATA string = path.Join(constants.BITFINEX_PATH, "pairs_list.json")
var BITFINEX_PAIRS_DETAILS string = path.Join(constants.BITFINEX_PATH, "pairs_info.json")
var BITFINEX_ORDERBOOK_DATA string = path.Join(constants.BITFINEX_PATH, "orders/")

type Bitfinex struct {
	PairsNames []string                                   `json:"pairs_name"`
	Pairs      map[string]datastructure.BitfinexPair      `json:"pairs_info"`
	OrderBook  map[string]datastructure.BitfinexOrderBook `json:"orderbook"`
	MakerFee   float32                                    `json:"maker_fee"`
	TakerFees  float32                                    `json:"taker_fee"`
	// FeePercent is delegated to save if the fee is in percent or in coin
	FeePercent bool `json:"fee_percent"`
}

func (b *Bitfinex) Init() {
	b.Pairs = make(map[string]datastructure.BitfinexPair)
	b.OrderBook = make(map[string]datastructure.BitfinexOrderBook)
}

// GetPairsList is delegated to retrieve the type of pairs in the Bitfinex market
func (b *Bitfinex) GetPairsList() error {
	var request req.Request
	var pairs []string
	var data []byte
	var err error

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(BITFINEX_PAIRS_DATA) {
		zap.S().Debugw("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(BITFINEX_PAIRS_DATA)
		if err != nil {
			zap.S().Debugw("Error reading data: " + err.Error())
			return err
		}
	} else {
		zap.S().Debugw("Sendind request to [" + BITFINEX_PAIRS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(BITFINEX_PAIRS_URL, "GET", nil, false)
		if resp.Error != nil {
			zap.S().Debugw("Error during http request. Err: " + resp.Error.Error())
			return err
		}
		if resp.StatusCode != 200 {
			zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode))
			return errors.New("STATUS_CODE_NOT_200")
		}
		data = resp.Body
	}

	err = json.Unmarshal(data, &pairs)
	if err != nil {
		zap.S().Warnw("Error during unmarshal! Err: " + err.Error())
		return err
	}

	b.PairsNames = pairs

	// Update the file with the new data
	utils.DumpStruct(pairs, BITFINEX_PAIRS_DATA)
	return nil
}

// GetPairsDetails is delegated to retrieve the information related to all pairs for execute order
func (b *Bitfinex) GetPairsDetails() error {
	var request req.Request
	var pairsInfo []datastructure.BitfinexPair
	var data []byte
	var err error

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(BITFINEX_PAIRS_DETAILS) {
		zap.S().Debugw("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(BITFINEX_PAIRS_DETAILS)
		if err != nil {
			zap.S().Warnw("Error reading data: " + err.Error())
			return err
		}
	} else {
		zap.S().Debugw("Sendind request to [" + BITFINEX_PAIRS_DETAILS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(BITFINEX_PAIRS_DETAILS_URL, "GET", nil, false)
		if resp.Error != nil {
			zap.S().Warnw("Error during http request. Err: " + resp.Error.Error())
			return resp.Error
		}
		if resp.StatusCode != 200 {
			zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode))
			return errors.New("NON_200_STATUS_CODE")
		}
		data = resp.Body
	}

	err = json.Unmarshal(data, &pairsInfo)

	if err != nil {
		zap.S().Warnw("Error during unmarshal! Err: " + err.Error())
		return err
	}

	for i := range pairsInfo {
		b.Pairs[pairsInfo[i].Pair] = pairsInfo[i]
	}

	// Update the file with the new data
	utils.DumpStruct(b.Pairs, BITFINEX_PAIRS_DETAILS)

	return nil
}

func (b *Bitfinex) GetAllOrderBook() error {
	var request req.Request
	var data []byte
	var err error

	for _, pair := range b.PairsNames {
		zap.S().Debugw("Managin pair: [" + pair + "]")
		if strings.Contains(pair, ":") {
			zap.S().Warnw("[" + pair + "] is not a tradable pair")
			continue
		}
		var orderbook datastructure.BitfinexOrderBook
		orderbook.Pair = pair
		file_data := path.Join(BITFINEX_ORDERBOOK_DATA, pair+".json")
		// Avoid to call the HTTP api if the data are present
		if fileutils.FileExists(file_data) {
			zap.S().Debugw("[" + pair + "] Data alredy present, avoiding to call the service")
			data, err = ioutil.ReadFile(file_data)
			if err != nil {
				zap.S().Warnw("Error reading data: " + err.Error())
				continue
			}
		} else {
			time.Sleep(2 * time.Second)
			url := BITFINEX_ORDER_BOOK_URL + pair
			zap.S().Debugw("Sendind request to [" + url + "]")
			// Call the HTTP method for retrieve the pairs
			resp := request.SendRequest(url, "GET", nil, false)
			if resp.Error != nil {
				zap.S().Warnw("Error during http request. Err: " + resp.Error.Error())
				// return resp.Error
				continue
			}
			if resp.StatusCode != 200 {
				zap.S().Warnw("Received a non 200 status code for pair [" + pair + "]: " + strconv.Itoa(resp.StatusCode) + " for pair [" + pair + "]")
				continue
			}
			data = resp.Body
		}

		err = json.Unmarshal(data, &orderbook)

		if err != nil {
			zap.S().Debugw("Error during unmarshal pair [" + pair + "]! Err: " + err.Error())
			continue
		}

		b.OrderBook[pair] = orderbook
		// Update the file with the new data
		utils.DumpStruct(b.OrderBook[pair], file_data)
	}

	// Update the file with the new data
	utils.DumpStruct(b.OrderBook, path.Join(constants.BITFINEX_PATH, "orders_all.json"))
	return nil
}

func (b *Bitfinex) SetFees() {
	b.MakerFee = 0.1
	b.TakerFees = 0.2
	b.FeePercent = false
}

func (b *Bitfinex) GetMarketData(pair string) (market.Market, error) {
	var markets market.Market
	markets.Asks = make(map[string][]market.MarketOrder, len(b.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(b.OrderBook))
	markets.MarketName = `BITFINEX`
	var order market.MarketOrder
	if orders, ok := b.OrderBook[pair]; ok {
		var asks []market.MarketOrder = make([]market.MarketOrder, len(orders.Asks))
		for i, ask := range orders.Asks {
			price, _ := strconv.ParseFloat(ask.Price, 64)
			volume, _ := strconv.ParseFloat(ask.Volume, 64)
			order.Price = price
			order.Volume = volume
			asks[i] = order
		}
		var bids []market.MarketOrder = make([]market.MarketOrder, len(orders.Bids))
		for i, bid := range orders.Bids {
			price, _ := strconv.ParseFloat(bid.Price, 64)
			volume, _ := strconv.ParseFloat(bid.Volume, 64)
			order.Price = price
			order.Volume = volume
			bids[i] = order
		}
		markets.Asks[pair] = asks
		markets.Bids[pair] = bids
		return markets, nil
	}
	return markets, errors.New("unable to find pair [" + pair + "]")
}
