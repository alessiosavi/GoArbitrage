package gemini

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/alessiosavi/GoArbitrage/datastructure/constants"
	datastructure "github.com/alessiosavi/GoArbitrage/datastructure/gemini"
	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	"github.com/alessiosavi/GoArbitrage/utils"
	fileutils "github.com/alessiosavi/GoGPUtils/files"
	req "github.com/alessiosavi/Requests"
)

// URL for understand the pairs traded in the market
const GEMINI_PAIRS_URL string = `https://api.sandbox.gemini.com/v1/symbols`
const GEMINI_ORDER_BOOK_URL string = `https://api.sandbox.gemini.com/v1/book/`

var GEMINI_PAIRS_DATA = path.Join(constants.GEMINI_PATH, "pairs_list.json")
var GEMINI_PAIRS_DETAILS = path.Join(constants.GEMINI_PATH, "pairs_info.json")
var GEMINI_ORDERBOOK_DATA = path.Join(constants.GEMINI_PATH, "orders/")

type Gemini struct {
	PairsNames []string                                 `json:"pairs_name"`
	OrderBook  map[string]datastructure.GeminiOrderBook `json:"orderbook"`
	PairsInfo  map[string]datastructure.GeminiPairs     `json:"pairs_info"`
	MakerFee   float64                                  `json:"maker_fee"`
	TakerFees  float64                                  `json:"taker_fee"`
	// FeePercent is delegated to save if the fee is in percent or in coin
	FeePercent bool `json:"fee_percent"`
}

func (g *Gemini) Init() {
	g.OrderBook = make(map[string]datastructure.GeminiOrderBook)
	g.PairsInfo = make(map[string]datastructure.GeminiPairs)
	g.SetFees()
}

// SetFees is delegated to initialize the fee type/amount for the given market
func (g *Gemini) SetFees() {
	g.MakerFee = 0.1
	g.TakerFees = 0.35
	g.FeePercent = true
}

// GetPairsList is delegated to retrieve the type of pairs in the Gemini market
func (g *Gemini) GetPairsList() error {
	var request req.Request
	var pairs []string
	var data []byte
	var err error

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(GEMINI_PAIRS_DATA) {
		zap.S().Debugw("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(GEMINI_PAIRS_DATA)
		if err != nil {
			zap.S().Warnw("Error reading data: " + err.Error())
			return err
		}

		err = json.Unmarshal(data, &g.PairsNames)

		if err != nil {
			zap.S().Warnw("Error during unmarshal! Err: " + err.Error())
			return nil
		}
		return nil

	} else {
		zap.S().Debugw("Sendind request to [" + GEMINI_PAIRS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(GEMINI_PAIRS_URL, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Second)
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

	err = json.Unmarshal(data, &pairs)

	if err != nil {
		zap.S().Warnw("Error during unmarshal! Err: " + err.Error())
		return err
	}

	g.PairsNames = pairs

	// Update the file with the new data
	utils.DumpStruct(pairs, GEMINI_PAIRS_DATA)
	return nil
}

// GetPairsDetails is delegated to read the file that contains the min order for the given pair
func (g *Gemini) GetPairsDetails() error {
	var err error
	var data []byte

	if !fileutils.FileExists(GEMINI_PAIRS_DETAILS) {
		zap.S().Warn("ERROR! File [" + GEMINI_PAIRS_DETAILS + "] not found!")
	}

	data, err = ioutil.ReadFile(GEMINI_PAIRS_DETAILS)
	if err != nil {
		zap.S().Warn("Error reading data: " + err.Error())
		return err
	}

	var pairs []datastructure.GeminiPairs
	err = json.Unmarshal(data, &pairs)
	if err != nil {
		zap.S().Warn("Error during unmarshal! Err: " + err.Error())
		return err
	}

	// Save pairs as a map
	for i := range pairs {
		g.PairsInfo[pairs[i].Pair] = pairs[i]
	}

	// Update the file with the new data
	utils.DumpStruct(g.PairsInfo, GEMINI_PAIRS_DETAILS+"_map")
	return nil
}

func (g *Gemini) GetAllOrderBook() error {
	var request req.Request
	var orders []datastructure.GeminiOrderBook
	var data []byte
	var err error

	for _, pair := range g.PairsNames {
		zap.S().Debugw("Managin pair: [" + pair + "]")
		var orderbook datastructure.GeminiOrderBook
		orderbook.Pair = pair
		file_data := path.Join(GEMINI_ORDERBOOK_DATA, pair+".json")
		// Avoid to call the HTTP api if the data are present
		if fileutils.FileExists(file_data) {
			zap.S().Debugw("[" + pair + "] Data alredy present, avoiding to call the service")
			data, err = ioutil.ReadFile(file_data)
			if err != nil {
				zap.S().Warnw("Error reading data: " + err.Error())
				continue
			}
		} else {
			// NOTE: limit the response to only 3 orders
			url := GEMINI_ORDER_BOOK_URL + pair + "?limit_bids=1&limit_asks=1"
			zap.S().Debugw("Sendind request to [" + url + "]")
			// Call the HTTP method for retrieve the pairs
			resp := request.SendRequest(url, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Second)
			if resp.Error != nil {
				zap.S().Warnw("Error during http request. Err: " + resp.Error.Error())
				continue
			}
			if resp.StatusCode != 200 {
				zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode) + " for pair [" + pair + "]")
				continue
			}
			data = resp.Body
		}

		err = json.Unmarshal(data, &orderbook)

		if err != nil {
			zap.S().Warnw("Error during unmarshal pair [" + pair + "]! Err: " + err.Error())
		} else {
			orders = append(orders, orderbook)
			// Update the file with the new data
			utils.DumpStruct(orderbook, file_data)
		}
	}

	for i := range orders {
		g.OrderBook[orders[i].Pair] = orders[i]
	}

	// Update the file with the new data
	utils.DumpStruct(g.OrderBook, path.Join(constants.GEMINI_PATH, "orders_all.json"))
	return nil
}

func (g *Gemini) GetMarketData(pair string) (market.Market, error) {
	var markets market.Market
	markets.Asks = make(map[string][]market.MarketOrder, len(g.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(g.OrderBook))
	markets.MarketName = `GEMINI`
	var order market.MarketOrder
	if orders, ok := g.OrderBook[pair]; ok {
		var asks = make([]market.MarketOrder, len(orders.Asks))
		for i, ask := range orders.Asks {
			price, _ := strconv.ParseFloat(ask.Price, 64)
			volume, _ := strconv.ParseFloat(ask.Volume, 64)
			order.Price = price
			order.Volume = volume
			asks[i] = order
		}
		var bids = make([]market.MarketOrder, len(orders.Bids))
		for i, bid := range orders.Bids {
			price, _ := strconv.ParseFloat(bid.Price, 64)
			volume, _ := strconv.ParseFloat(bid.Volume, 64)
			order.Price = price
			order.Volume = volume
			bids[i] = order
		}
		markets.Asks[pair] = asks
		markets.Bids[pair] = bids
		markets.MakerFee = g.MakerFee
		markets.TakerFee = g.TakerFees
		return markets, nil
	}
	return markets, errors.New("unable to find pair [" + pair + "]")
}

// GetMarketsData is delegated to convert the internal asks and bids struct to the common "market" struct
func (g *Gemini) GetMarketsData() market.Market {
	var markets market.Market
	// Standardize key for common coin
	var key_standard string
	markets.Asks = make(map[string][]market.MarketOrder, len(g.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(g.OrderBook))
	markets.MarketName = `GEMINI`
	// var i int
	var order market.MarketOrder
	for key := range g.OrderBook {
		key_standard = strings.Replace(strings.ToLower(key), "-", "", 1)
		var asks = make([]market.MarketOrder, len(g.OrderBook[key].Asks))
		for i, ask := range g.OrderBook[key].Asks {
			price, _ := strconv.ParseFloat(ask.Price, 64)
			volume, _ := strconv.ParseFloat(ask.Volume, 64)
			order.Price = price
			order.Volume = volume
			asks[i] = order
		}
		var bids = make([]market.MarketOrder, len(g.OrderBook[key].Bids))
		for i, bid := range g.OrderBook[key].Bids {
			price, _ := strconv.ParseFloat(bid.Price, 64)
			volume, _ := strconv.ParseFloat(bid.Volume, 64)
			order.Price = price
			order.Volume = volume
			bids[i] = order
		}
		markets.Asks[key_standard] = asks
		markets.Bids[key_standard] = bids
	}

	return markets
}

func (g *Gemini) GetOrderBook(pair string) error {
	var request req.Request
	var order datastructure.GeminiOrderBook
	var data []byte
	var err error

	// NOTE: limit the response to only 3 orders
	url := GEMINI_ORDER_BOOK_URL + pair + "?limit_bids=1&limit_asks=1"
	zap.S().Debugw("Sendind request to [" + url + "]")
	// Call the HTTP method for retrieve the pairs
	resp := request.SendRequest(url, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Second)
	if resp.Error != nil {
		zap.S().Warnw("Error during http request. Err: " + resp.Error.Error())
		return err
	}

	if resp.StatusCode != 200 {
		zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode) + " for pair [" + pair + "]")
		return errors.New("NOT_200_HTTP_STATUS")
	}
	data = resp.Body

	err = json.Unmarshal(data, &order)
	if err != nil {
		zap.S().Warnw("Error during unmarshal pair [" + pair + "]! Err: " + err.Error())
		return err
	}

	if len(g.OrderBook) == 0 {
		g.OrderBook = make(map[string]datastructure.GeminiOrderBook)
	}
	// Update the file with the new data
	//utils.DumpStruct(order, path.Join(GEMINI_ORDERBOOK_DATA, pair+".json"))

	// Update the file with the new data
	return nil
}
