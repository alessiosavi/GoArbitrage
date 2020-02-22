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
	"github.com/alessiosavi/GoArbitrage/datastructure/constants"
	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	fileutils "github.com/alessiosavi/GoGPUtils/files"
	req "github.com/alessiosavi/Requests"
)

const BITFINEX_PAIRS_URL string = `https://api.bitfinex.com/v1/symbols`
const BITFINEX_PAIRS_DETAILS_URL string = `https://api.bitfinex.com/v1/symbols_details`
const BITFINEX_ORDER_BOOK_URL string = `https://api.bitfinex.com/v1/book/`

const BITFINEX_TICKERS_DETAILS string = `https://api-pub.bitfinex.com/v2/conf/pub:map:currency:sym`

var BITFINEX_PAIRS_DATA = path.Join(constants.BITFINEX_PATH, "pairs_list.json")
var BITFINEX_PAIRS_DETAILS = path.Join(constants.BITFINEX_PATH, "pairs_info.json")
var BITFINEX_ORDERBOOK_DATA = path.Join(constants.BITFINEX_PATH, "orders/")

type BtfinexTickers [][][]string

type Bitfinex struct {
	PairsNames []string                                   `json:"pairs_name"`
	Pairs      map[string]datastructure.BitfinexPair      `json:"pairs_info"`
	OrderBook  map[string]datastructure.BitfinexOrderBook `json:"orderbook"`
	MakerFee   float64                                    `json:"maker_fee"`
	TakerFees  float64                                    `json:"taker_fee"`
	// FeePercent is delegated to save if the fee is in percent or in coin
	FeePercent bool `json:"fee_percent"`
	Tickers    []string
}

// Init is delegated to initialize the map for the given market
func (b *Bitfinex) Init() {
	b.Pairs = make(map[string]datastructure.BitfinexPair)
	b.OrderBook = make(map[string]datastructure.BitfinexOrderBook)
	b.SetFees()
}

// SetFees is delegated to initialize the fee type/amount for the given market
func (b *Bitfinex) SetFees() {
	b.MakerFee = 0.1
	b.TakerFees = 0.2
	b.FeePercent = true
}

// GetTickers is delegated to retrive the list of tickers tradable on Bitfinex
func (b *Bitfinex) GetTickers() error {
	var (
		request req.Request
		data    []byte
		err     error
		tickers BtfinexTickers
	)
	zap.S().Debugw("Sending request to [" + BITFINEX_TICKERS_DETAILS + "]")
	// Call the HTTP method for retrieve the pairs
	resp := request.SendRequest(BITFINEX_TICKERS_DETAILS, "GET", nil, nil, false, 10*time.Second)
	if resp.Error != nil {
		zap.S().Warnw("Error during http request. Err: " + resp.Error.Error())
		return resp.Error
	}
	if resp.StatusCode != 200 {
		zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode))
		return errors.New("NON_200_STATUS_CODE")
	}
	data = resp.Body

	if err = json.Unmarshal(data, &tickers); err != nil {
		zap.S().Warnf("Error: %s", err.Error())
		zap.S().Infof("Data: %s", string(data))
		return err
	}
	var t = make([]string, len(tickers[0]))
	for i := range tickers[0] {
		t[i] = tickers[0][i][0]
	}
	b.Tickers = t
	return nil
}

// GetPairsList is delegated to retrieve the type of pairs in the Bitfinex market
func (b *Bitfinex) GetPairsList() error {
	var (
		request req.Request
		pairs   []string
		data    []byte
		err     error
	)

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(BITFINEX_PAIRS_DATA) {
		zap.S().Debugw("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(BITFINEX_PAIRS_DATA)
		if err != nil {
			zap.S().Debugw("Error reading data: " + err.Error())
			return err
		}
	} else {
		zap.S().Debugw("Sending request to [" + BITFINEX_PAIRS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(BITFINEX_PAIRS_URL, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Second)
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
		zap.S().Debugw("Sending request to [" + BITFINEX_PAIRS_DETAILS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(BITFINEX_PAIRS_DETAILS_URL, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Second)
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

// GetAllOrderBook is delegated to retrieve the order book for all the currencies
func (b *Bitfinex) GetAllOrderBook() error {
	var request req.Request
	var data []byte
	var err error

	for _, pair := range b.PairsNames {
		zap.S().Debugw("Managin pair: [" + pair + "]")
		if strings.Contains(pair, ":") {
			zap.S().Info("[" + pair + "] is not a tradable pair")
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
			url := BITFINEX_ORDER_BOOK_URL + pair + "?limit_bids=1&limit_asks=1"
			zap.S().Debugw("Sending request to [" + url + "]")
			// Call the HTTP method for retrieve the pairs
			resp := request.SendRequest(url, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Second)
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

// GetMarketData is delegated to convert the order book into a standard `market` struct
func (b *Bitfinex) GetMarketData(pair string) (market.Market, error) {
	var markets market.Market
	markets.Asks = make(map[string][]market.MarketOrder, len(b.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(b.OrderBook))
	markets.MarketName = `BITFINEX`
	minVolume, _ := strconv.ParseFloat(b.Pairs[pair].MinOrder, 64)
	var order market.MarketOrder
	if orders, ok := b.OrderBook[pair]; ok {
		var asks = make([]market.MarketOrder, len(orders.Asks))
		for i, ask := range orders.Asks {
			price, _ := strconv.ParseFloat(ask.Price, 64)
			volume, _ := strconv.ParseFloat(ask.Volume, 64)
			order.Price = price
			order.Volume = volume
			order.MinVolume = minVolume
			asks[i] = order
		}
		var bids = make([]market.MarketOrder, len(orders.Bids))
		for i, bid := range orders.Bids {
			price, _ := strconv.ParseFloat(bid.Price, 64)
			volume, _ := strconv.ParseFloat(bid.Volume, 64)
			order.Price = price
			order.Volume = volume
			order.MinVolume = minVolume
			bids[i] = order
		}
		markets.Asks[pair] = asks
		markets.Bids[pair] = bids
		markets.MakerFee = b.MakerFee
		markets.TakerFee = b.TakerFees
		return markets, nil
	}
	return markets, errors.New("unable to find pair [" + pair + "]")
}

// GetMarketsData is delegated to convert the internal asks and bids struct to the common "market" struct
func (b *Bitfinex) GetMarketsData() market.Market {
	var markets market.Market
	// Standardize key for common coin
	var key_standard string
	markets.Asks = make(map[string][]market.MarketOrder, len(b.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(b.OrderBook))

	markets.MarketName = `BITFINEX`

	var order market.MarketOrder
	for key := range b.OrderBook {

		key_standard = strings.Replace(strings.ToLower(key), "-", "", 1)
		var asks = make([]market.MarketOrder, len(b.OrderBook[key].Asks))
		for i, ask := range b.OrderBook[key].Asks {
			price, _ := strconv.ParseFloat(ask.Price, 64)
			volume, _ := strconv.ParseFloat(ask.Volume, 64)
			order.Price = price
			order.Volume = volume
			asks[i] = order
		}
		var bids = make([]market.MarketOrder, len(b.OrderBook[key].Bids))
		for i, bid := range b.OrderBook[key].Bids {
			price, _ := strconv.ParseFloat(bid.Price, 64)
			volume, _ := strconv.ParseFloat(bid.Volume, 64)
			order.Price = price
			order.Volume = volume
			bids[i] = order
		}
		//markets.MinPrice
		markets.Asks[key_standard] = asks
		markets.Bids[key_standard] = bids
	}

	return markets
}

// GetOrderBook is delegated to retrieve the order book for the given pair
func (b *Bitfinex) GetOrderBook(pair string) error {
	var request req.Request
	var data []byte
	var orderbook datastructure.BitfinexOrderBook
	var err error

	zap.S().Debugw("Managin pair: [" + pair + "]")
	if strings.Contains(pair, ":") {
		zap.S().Info("[" + pair + "] is not a tradable pair")
		return errors.New("PAIRS_NOT_TRADABLE")
	}
	orderbook.Pair = pair

	url := BITFINEX_ORDER_BOOK_URL + pair + "?limit_bids=1&limit_asks=1"
	zap.S().Debugw("Sending request to [" + url + "]")
	// Call the HTTP method for retrieve the pairs
	resp := request.SendRequest(url, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Second)
	if resp.Error != nil {
		zap.S().Warnw("Error during http request. Err: " + resp.Error.Error())
		return resp.Error
	}
	if resp.StatusCode != 200 {
		zap.S().Warnw("Received a non 200 status code for pair [" + pair + "]: " + strconv.Itoa(resp.StatusCode) + " for pair [" + pair + "]")
		return errors.New("NOT_200_HTTP_STATUS")
	}
	data = resp.Body

	err = json.Unmarshal(data, &orderbook)

	if err != nil {
		zap.S().Debugw("Error during unmarshal pair [" + pair + "]! Err: " + err.Error())
		return err
	}

	if len(b.OrderBook) == 0 {
		b.OrderBook = make(map[string]datastructure.BitfinexOrderBook)
	}
	b.OrderBook[pair] = orderbook
	// Update the file with the new data
	// utils.DumpStruct(b.OrderBook[pair], path.Join(BITFINEX_ORDERBOOK_DATA, pair+".json"))

	return nil
}
