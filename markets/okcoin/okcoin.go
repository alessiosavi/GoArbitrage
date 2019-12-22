package okcoin

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/alessiosavi/GoArbitrage/utils"
	"go.uber.org/zap"

	constants "github.com/alessiosavi/GoArbitrage/datastructure/constants"
	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	datastructure "github.com/alessiosavi/GoArbitrage/datastructure/okcoin"
	fileutils "github.com/alessiosavi/GoGPUtils/files"
	req "github.com/alessiosavi/Requests"
)

// `https://www.okcoin.com/api/spot/v3/instruments`

const OKCOIN_PAIRS_URL string = `https://www.okcoin.com/api/spot/v3/instruments/ticker`
const OKCOIN_PAIRS_DETAILS_URL string = `https://www.okcoin.com/api/spot/v3/instruments/`

var OKCOIN_PAIRS_DATA string = path.Join(constants.OKCOIN_PATH, "pairs_list.json")
var OKCOIN_PAIRS_DETAILS string = path.Join(constants.OKCOIN_PATH, "pairs_info.json")
var OKCOIN_ORDERBOOK_DATA string = path.Join(constants.OKCOIN_PATH, "orders/")

type OkCoin struct {
	PairsName []string                                 `json:"pairs_name"`
	Pairs     map[string]datastructure.OkCoinPairs     `json:"pairs"`
	OrderBook map[string]datastructure.OkCoinOrderBook `json:"orderbook"`
	MakerFee  float32                                  `json:"maker_fee"`
	TakerFees float32                                  `json:"taker_fee"`
	// FeePercent is delegated to save if the fee is in percent or in coin
	FeePercent bool `json:"fee_percent"`
}

func (o *OkCoin) Init() {
	o.Pairs = make(map[string]datastructure.OkCoinPairs)
	o.OrderBook = make(map[string]datastructure.OkCoinOrderBook)
}

func (o *OkCoin) GetMarketData(pair string) (market.Market, error) {
	var markets market.Market
	markets.Asks = make(map[string][]market.MarketOrder, len(o.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(o.OrderBook))
	markets.MarketName = `OKCOIN`
	var order market.MarketOrder
	if orders, ok := o.OrderBook[pair]; ok {
		var asks []market.MarketOrder = make([]market.MarketOrder, len(orders.Asks))
		for i, ask := range orders.Asks {
			price, _ := strconv.ParseFloat(ask[0], 64)
			volume, _ := strconv.ParseFloat(ask[1], 64)
			order.Price = price
			order.Volume = volume
			asks[i] = order
		}
		var bids []market.MarketOrder = make([]market.MarketOrder, len(orders.Bids))
		for i, bid := range orders.Bids {
			price, _ := strconv.ParseFloat(bid[0], 64)
			volume, _ := strconv.ParseFloat(bid[1], 64)
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

// GetMarketsData is delegated to convert the internal asks and bids struct to the common "market" struct
func (o *OkCoin) GetMarketsData() market.Market {
	var markets market.Market
	// Standardize key for common coin
	var key_standard string
	markets.Asks = make(map[string][]market.MarketOrder, len(o.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(o.OrderBook))
	markets.MarketName = `OKCOIN`
	// var i int
	var order market.MarketOrder
	for key := range o.OrderBook {
		key_standard = strings.Replace(strings.ToLower(key), "-", "", 1)
		var asks []market.MarketOrder = make([]market.MarketOrder, len(o.OrderBook[key].Asks))
		for i, ask := range o.OrderBook[key].Asks {
			price, _ := strconv.ParseFloat(ask[0], 64)
			volume, _ := strconv.ParseFloat(ask[1], 64)
			order.Price = price
			order.Volume = volume
			asks[i] = order
		}
		var bids []market.MarketOrder = make([]market.MarketOrder, len(o.OrderBook[key].Bids))
		for i, bid := range o.OrderBook[key].Bids {
			price, _ := strconv.ParseFloat(bid[0], 64)
			volume, _ := strconv.ParseFloat(bid[1], 64)
			order.Price = price
			order.Volume = volume
			bids[i] = order
		}
		markets.Asks[key_standard] = asks
		markets.Bids[key_standard] = bids
	}

	return markets
}

// GetPairsList is delegated to retrieve the type of pairs in the Bitfinex market
func (ok *OkCoin) GetPairsList() error {
	var request req.Request
	var data []byte
	var err error
	var pairs_raw []datastructure.OkCoinPair

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(OKCOIN_PAIRS_DATA) {
		zap.S().Debugw("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(OKCOIN_PAIRS_DATA)
		if err != nil {
			zap.S().Warnw("Error reading data: " + err.Error())
			return err
		}

		err = json.Unmarshal(data, &ok.PairsName)

		if err != nil {
			zap.S().Warnw("Error during unmarshal! Err: " + err.Error())
			return err
		}
		return nil

	} else {
		zap.S().Debugw("Sendind request to [" + OKCOIN_PAIRS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(OKCOIN_PAIRS_URL, "GET", nil, false)
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
	err = json.Unmarshal(data, &pairs_raw)

	if err != nil {
		zap.S().Warnw("Error during unmarshal! Err: " + err.Error())
		return err
	}
	ok.PairsName = make([]string, len(pairs_raw))
	for i := range pairs_raw {
		ok.PairsName[i] = pairs_raw[i].ProductID
	}
	// Update the file with the new data
	utils.DumpStruct(ok.PairsName, OKCOIN_PAIRS_DATA)
	return nil
}

// GetPairsDetails is delegated to retrieve the information related to the pairs
func (ok *OkCoin) GetPairsDetails() error {
	var request req.Request
	var pairsInfo []datastructure.OkCoinPairs
	var data []byte
	var err error

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(OKCOIN_PAIRS_DETAILS) {
		zap.S().Debugw("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(OKCOIN_PAIRS_DETAILS)
		if err != nil {
			zap.S().Debugw("Error reading data: " + err.Error())
			return err
		}
	} else {
		zap.S().Debugw("Sendind request to [" + OKCOIN_PAIRS_DETAILS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(OKCOIN_PAIRS_DETAILS_URL, "GET", nil, false)
		if resp.Error != nil {
			zap.S().Debugw("Error during http request. Err: " + resp.Error.Error())
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
		zap.S().Debugw("Error during unmarshal! Err: " + err.Error())
		return err
	}

	ok.Pairs = make(map[string]datastructure.OkCoinPairs, len(pairsInfo))

	for i := range pairsInfo {
		pairsInfo[i].Pair = pairsInfo[i].BaseCurrency + "-" + pairsInfo[i].QuoteCurrency
		ok.Pairs[pairsInfo[i].Pair] = pairsInfo[i]
	}

	// Update the file with the new data
	utils.DumpStruct(pairsInfo, OKCOIN_PAIRS_DETAILS)
	return nil
}

func (ok *OkCoin) GetOrderBook(pair string) error {
	var request req.Request
	var order datastructure.OkCoinOrderBook
	var data []byte
	var err error

	url := OKCOIN_PAIRS_DETAILS_URL + pair + "/book?size=1"
	zap.S().Debug("Sendind request to [" + url + "]")
	// Call the HTTP method for retrieve the pairs
	resp := request.SendRequest(url, "GET", nil, false)
	if resp.Error != nil {
		zap.S().Debug("Error during http request. Err: " + resp.Error.Error())
		return resp.Error
	}
	if resp.StatusCode != 200 {
		zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode) + " for pair [" + pair + "]")
		log.Println("Request -> ", request)
		log.Println("Response -> ", resp.Dump())
		return errors.New("NOT_200_HTTP_STATUS")
	}
	data = resp.Body

	err = json.Unmarshal(data, &order)
	if err != nil {
		zap.S().Debugw("Error during unmarshal! Err: " + err.Error())
		return err
	}

	if len(ok.OrderBook) == 0 {
		ok.OrderBook = make(map[string]datastructure.OkCoinOrderBook)
	}
	ok.OrderBook[pair] = order

	// Update the file with the new data
	//utils.DumpStruct(order, path.Join(OKCOIN_ORDERBOOK_DATA, pair+".json"))
	return nil
}

func (ok *OkCoin) GetAllOrderBook() error {
	var request req.Request
	var orders map[string]datastructure.OkCoinOrderBook = make(map[string]datastructure.OkCoinOrderBook, len(ok.PairsName))
	var data []byte
	var err error

	for _, pair := range ok.PairsName {
		zap.S().Debugw("Managin pair: [" + pair + "]")
		if strings.Contains(pair, ":") {
			zap.S().Warnw("[" + pair + "] is not a tradable pair")
			continue
		}
		var orderbook datastructure.OkCoinOrderBook
		orderbook.Pair = pair
		file_data := path.Join(OKCOIN_ORDERBOOK_DATA, pair+".json")
		// Avoid to call the HTTP api if the data are present
		if fileutils.FileExists(file_data) {
			zap.S().Debugw("[" + pair + "] Data alredy present, avoiding to call the service")
			data, err = ioutil.ReadFile(file_data)
			if err != nil {
				zap.S().Debugw("Error reading data: " + err.Error())
				continue
			}
		} else {
			time.Sleep(100 * time.Millisecond)
			url := OKCOIN_PAIRS_DETAILS_URL + pair + "/book?size=1"
			zap.S().Debugw("Sendind request to [" + url + "]")
			// Call the HTTP method for retrieve the pairs
			resp := request.SendRequest(url, "GET", nil, false)
			if resp.Error != nil {
				zap.S().Debugw("Error during http request. Err: " + resp.Error.Error())
				continue
			}
			if resp.StatusCode != 200 {
				zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode) + " for pair [" + pair + "]")
				log.Println("Request -> ", request)
				log.Println("Response -> ", resp.Dump())
				continue
			}
			data = resp.Body

		}

		err = json.Unmarshal(data, &orderbook)
		if err != nil {
			zap.S().Debugw("Error during unmarshal! Err: " + err.Error())
		} else {
			orders[pair] = orderbook
			// Update the file with the new data
			utils.DumpStruct(orderbook, file_data)
		}
	}

	ok.OrderBook = orders

	// Update the file with the new data
	utils.DumpStruct(ok.OrderBook, path.Join(constants.OKCOIN_PATH, "orders_all.json"))

	return nil
}

// getBookUrl is delegated to generate the URL for the given pairs
func getBookUrl(pairs string, size int, depth float64) string {
	u, err := url.Parse(OKCOIN_PAIRS_DETAILS_URL)
	if err != nil {
		zap.S().Warnw("Error creating URL!")
		return ""
	}
	u.Path = path.Join(u.Path, pairs)

	url := u.String() + "/book?size=" + strconv.Itoa(size)
	if depth != 0 {
		url += "&depth=" + strconv.FormatFloat(depth, 'f', -1, 64)
	}

	return url
}

var allowed_base []string = []string{"EUR", "EURS", "USD", "USDT", "SGD"}

// ParsePair is delegated to convert the given pair into the pair compliant with kraken
func (o *OkCoin) ParsePair(pair string) string {

	// Expected PAIR-BASE => adausd --> ADA-USD

	if strings.Contains(pair, "-") {
		return pair
	}

	// 1 Extract the last 3 char from the string

	lastchars := strings.ToUpper(pair[len(pair)-3:])

	var newpair string

	// 2 Switch case for verify if the coin is in the base allowed

	for i := range allowed_base {
		if lastchars == allowed_base[i] {
			// 3 If is a 3char pair, than add a `-` just before the 3 char
			newpair = strings.ToUpper(pair[:len(pair)-3] + "-" + lastchars)
			return newpair
		}
	}

	// 4 If not, then take 4 char and add a `-` just before the 4 char
	newpair = strings.ToUpper(pair[:len(pair)-4] + "-" + pair[len(pair)-4:])

	return newpair
}
