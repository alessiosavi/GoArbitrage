package kraken

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/alessiosavi/GoArbitrage/datastructure/constants"
	datastructure "github.com/alessiosavi/GoArbitrage/datastructure/kraken"
	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	"github.com/alessiosavi/GoArbitrage/utils"
	fileutils "github.com/alessiosavi/GoGPUtils/files"

	req "github.com/alessiosavi/Requests"
)

type Kraken struct {
	PairsNames []string                                 `json:"pairs_name"`
	Pairs      map[string]datastructure.KrakenPair      `json:"pairs"`
	OrderBook  map[string]datastructure.KrakenOrderBook `json:"orderbook"`
	MakerFee   float64                                  `json:"maker_fee"`
	TakerFees  float64                                  `json:"taker_fee"`
	// FeePercent is delegated to save if the fee is in percent or in coin
	FeePercent bool `json:"fee_percent"`
	Tickers    []string
}

const KRAKEN_TICKERS_URL string = `https://api.kraken.com/0/public/Assets`
const KRAKEN_PAIRS_DETAILS_URL string = `https://api.kraken.com/0/public/AssetPairs`
const KRAKEN_ORDER_BOOK_URL string = `https://api.kraken.com/0/public/Depth?pair=`

var KRAKEN_PAIRS_DATA = path.Join(constants.KRAKEN_PATH, "pairs_list.json")
var KRAKEN_PAIRS_DETAILS = path.Join(constants.KRAKEN_PATH, "pairs_info.json")
var KRAKEN_ORDERBOOK_DATA = path.Join(constants.KRAKEN_PATH, "orders/")

// Init is delegated to initialize the maps for the kraken
func (k *Kraken) Init() {
	k.Pairs = make(map[string]datastructure.KrakenPair)
	k.OrderBook = make(map[string]datastructure.KrakenOrderBook)
	k.SetFees()
}

// SetFees is delegated to initialize the fee type/amount for the given market
func (k *Kraken) SetFees() {
	k.MakerFee = 0.16
	k.TakerFees = 0.26
	k.FeePercent = true
}

// GetTickers is delegated to retrieve the list of tickers tradable in the exchange
func (k *Kraken) GetTickers() error {
	res := datastructure.Tickers{}
	var err error
	var request req.Request
	var data []byte
	var tickers []string
	resp := request.SendRequest(KRAKEN_TICKERS_URL, "GET", nil, nil, false, 10*time.Second)
	if resp.Error != nil {
		zap.S().Debugw("Error during http request. Err: " + resp.Error.Error())
		return resp.Error
	}
	if resp.StatusCode != 200 {
		zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode))
		return errors.New("NON_200_STATUS_CODE")
	}
	data = resp.Body
	if err = json.Unmarshal(data, &res); err != nil {
		zap.S().Warn("ERROR! :" + err.Error())
		return err
	}
	zap.S().Infof("Data: %v", res.Result)
	tickers = make([]string, len(res.Result))
	i := 0
	for key := range res.Result {
		tickers[i] = res.Result[key].Altname
		i++
	}
	k.Tickers = tickers
	return nil
}

// GetPairsDetails is delegated to retrieve the pairs detail and the pairs names
func (k *Kraken) GetPairsDetails() error {
	var request req.Request
	var data []byte
	var err error

	// Avoid to call the HTTP api if the data are present
	if fileutils.FileExists(KRAKEN_PAIRS_DETAILS) {
		zap.S().Debugw("Data alredy present, avoiding to call the service")
		data, err = ioutil.ReadFile(KRAKEN_PAIRS_DETAILS)
		if err != nil {
			zap.S().Warnw("Error reading data: " + err.Error())
			return err
		}
		err = json.Unmarshal(data, &k.Pairs)
		if err != nil {
			zap.S().Warnw("Error reading data: " + err.Error())
			return err
		}

		k.PairsNames = make([]string, len(k.Pairs))
		var i int
		for key := range k.Pairs {
			k.PairsNames[i] = k.Pairs[key].Altname
			i++
		}
		return nil
	}

	zap.S().Debugw("Sendind request to [" + KRAKEN_PAIRS_DETAILS_URL + "]")
	// Call the HTTP method for retrieve the pairs
	resp := request.SendRequest(KRAKEN_PAIRS_DETAILS_URL, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Millisecond)
	if resp.Error != nil {
		zap.S().Debugw("Error during http request. Err: " + resp.Error.Error())
		return resp.Error
	}
	if resp.StatusCode != 200 {
		zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode))
		return errors.New("NON_200_STATUS_CODE")
	}
	data = resp.Body

	k.Pairs = loadKrakenPairs(data)
	if k.Pairs == nil {
		err = errors.New("UNABLE_LOAD_PAIRS")
		zap.S().Errorw("Unable to load kraken pairs!")
		return err
	}

	k.PairsNames = make([]string, len(k.Pairs))
	var i int
	for key := range k.Pairs {
		k.PairsNames[i] = k.Pairs[key].Altname
		i++
	}
	utils.DumpStruct(k.Pairs, KRAKEN_PAIRS_DETAILS)
	return nil
}

// GetAllOrderBook is delegated to download all the order book related to the pair traded
func (k *Kraken) GetAllOrderBook() error {
	var request req.Request
	var err error
	var data []byte

	for _, pair := range k.PairsNames {
		var order datastructure.KrakenOrderBook
		zap.S().Debugw("Managin pair: [" + pair + "]")
		file := path.Join(KRAKEN_ORDERBOOK_DATA, pair+".json")
		if fileutils.FileExists(file) {
			zap.S().Debug("Data [" + pair + "] alredy present, avoiding to call the service")
			data, err = ioutil.ReadFile(file)
			if err != nil {
				zap.S().Warn("Error reading data: " + err.Error())
				continue
			}
			err = json.Unmarshal(data, &order)
			//order, err = loadOrderBook(data)
			if err != nil {
				zap.S().Warn("Error reading data: " + err.Error())
				continue
			} else {
				//order.Pair = pair
				k.OrderBook[pair] = order
			}
		} else {
			url := KRAKEN_ORDER_BOOK_URL + pair + `&count=1`
			zap.S().Debugw("Sendind request to [" + url + "]")
			// Call the HTTP method for retrieve the pairs
			resp := request.SendRequest(url, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Millisecond)
			if resp.Error != nil {
				zap.S().Warnw("Error during http request. Err: " + resp.Error.Error())
				continue
			}
			if resp.StatusCode != 200 {
				zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode))
				continue
			}
			data = resp.Body

			order, err = loadOrderBook(data)
			if err != nil {
				zap.S().Warnw("Error during retrieve of pair [" + pair + "]!")
			} else {
				order.Pair = pair
				k.OrderBook[pair] = order
				utils.DumpStruct(order, file)
			}
		}
	}

	utils.DumpStruct(k.OrderBook, path.Join(constants.KRAKEN_PATH, "orders_all.json"))
	return nil
}

func loadOrderBook(data []byte) (datastructure.KrakenOrderBook, error) {
	res := &datastructure.Response{}
	var err error
	if err = json.Unmarshal(data, res); err != nil {
		zap.S().Warn("ERROR! :" + err.Error())
		return datastructure.KrakenOrderBook{}, err
	}
	// Will have only 1 key
	for key /*,value*/ := range res.Result {

		// fmt.Println(key)
		// for i, ask := range value.Asks {
		// 	log.Println(fmt.Sprintf("Asks[%d] = %#v\n", i, ask))
		// }
		// for i, bid := range value.Bids {
		// 	log.Println(fmt.Sprintf("Bids[%d] = %#v\n", i, bid))
		// }
		return res.Result[key], nil
	}
	return datastructure.KrakenOrderBook{}, err
}

// loadKrakenPairs is delegated to load the data related to pairs info
func loadKrakenPairs(data []byte) map[string]datastructure.KrakenPair {
	var err error

	// Creating the maps for the JSON data
	m := map[string]interface{}{}
	var pairInfo = make(map[string]datastructure.KrakenPair)

	// Parsing/Unmarshalling JSON
	err = json.Unmarshal(data, &m)

	if err != nil {
		zap.S().Debugw("Error unmarshalling data: " + err.Error())
		return nil
	}

	a := reflect.ValueOf(m["result"])

	if a.Kind() == reflect.Map {
		for _, key := range a.MapKeys() {
			strct := a.MapIndex(key)
			m, _ := strct.Interface().(map[string]interface{})
			data, err := json.Marshal(m)
			if err != nil {
				zap.S().Warnw("Panic on key: ", key, " ERR: "+err.Error())
				continue
			}
			var result datastructure.KrakenPair
			err = json.Unmarshal(data, &result)
			if err != nil {
				zap.S().Warnw("Panic on key: ", key, " during unmarshal. ERR: "+err.Error())
				continue
			}
			pairInfo[key.String()] = result
		}
	}
	return pairInfo
}

// GetMarketData is delegated to convert the order book into a standard `market` struct
func (k *Kraken) GetMarketData(pair string) (market.Market, error) {
	var markets market.Market
	markets.Asks = make(map[string][]market.MarketOrder, len(k.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(k.OrderBook))
	markets.MarketName = `KRAKEN`
	var order market.MarketOrder
	if orders, ok := k.OrderBook[pair]; ok {
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
		return markets, nil
	}
	return markets, errors.New("unable to find pair [" + pair + "]")
}

// GetMarketsData is delegated to convert the internal asks and bids struct to the common "market" struct
func (k *Kraken) GetMarketsData() market.Market {
	var markets market.Market
	// Standardize key for common coin
	var key_standard string
	markets.Asks = make(map[string][]market.MarketOrder, len(k.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(k.OrderBook))
	markets.MarketName = `KRAKEN`
	markets.MakerFee = k.MakerFee
	markets.TakerFee = k.TakerFees
	// var i int
	var order market.MarketOrder
	amounts := utils.LoadMinAmountKraken(path.Join(constants.KRAKEN_PATH, "min_amount.txt"))

	for key := range k.OrderBook {
		key_standard = strings.Replace(strings.ToLower(key), "-", "", 1)
		var asks = make([]market.MarketOrder, len(k.OrderBook[key].Asks))

		for i, ask := range k.OrderBook[key].Asks {
			price, _ := strconv.ParseFloat(ask.Price, 64)
			volume, _ := strconv.ParseFloat(ask.Volume, 64)
			order.Price = price
			order.Volume = volume
			asks[i] = order
		}
		var bids = make([]market.MarketOrder, len(k.OrderBook[key].Bids))
		for i, bid := range k.OrderBook[key].Bids {
			price, _ := strconv.ParseFloat(bid.Price, 64)
			volume, _ := strconv.ParseFloat(bid.Volume, 64)
			order.Price = price
			order.Volume = volume
			bids[i] = order
		}
		markets.Asks[key_standard] = asks
		markets.Bids[key_standard] = bids

		for key := range amounts {
			if strings.HasPrefix(key_standard, key) {
				for i := range markets.Asks[key_standard] {
					markets.Asks[key_standard][i].MinVolume = amounts[key]
					markets.Bids[key_standard][i].MinVolume = amounts[key]
				}
				break
			}
		}
	}

	return markets
}

func (k *Kraken) GetOrderBook(pair string) error {
	var request req.Request
	var err error
	var data []byte

	var order datastructure.KrakenOrderBook
	zap.S().Debugw("Managin pair: [" + pair + "]")
	url := KRAKEN_ORDER_BOOK_URL + pair + `&count=1`
	zap.S().Debugw("Sendind request to [" + url + "]")
	// Call the HTTP method for retrieve the pairs
	resp := request.SendRequest(url, "GET", nil, nil, false, constants.TIMEOUT_REQ*time.Millisecond)
	if resp.Error != nil {
		zap.S().Warnw("Error during http request. Err: " + resp.Error.Error())
		return err
	}
	if resp.StatusCode != 200 {
		zap.S().Warnw("Received a non 200 status code: " + strconv.Itoa(resp.StatusCode))
		return errors.New("NOT_200_HTTP_STATUS")
	}
	data = resp.Body

	order, err = loadOrderBook(data)
	if err != nil {
		zap.S().Warnw("Error during retrieve of pair [" + pair + "]!")
		return err
	}
	order.Pair = pair
	if len(k.OrderBook) == 0 {
		k.OrderBook = map[string]datastructure.KrakenOrderBook{}
	}
	k.OrderBook[pair] = order
	//utils.DumpStruct(order, path.Join(KRAKEN_ORDERBOOK_DATA, pair+".json"))
	return nil
}

// ParsePair is delegated to convert the given pair into the pair compliant with kraken
func (k *Kraken) ParsePair(pair string) string {
	return strings.ToUpper(pair)
}
