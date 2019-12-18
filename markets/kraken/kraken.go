package kraken

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"path"
	"reflect"
	"strconv"

	"go.uber.org/zap"

	constants "github.com/alessiosavi/GoArbitrage/datastructure/constants"
	datastructure "github.com/alessiosavi/GoArbitrage/datastructure/kraken"
	"github.com/alessiosavi/GoArbitrage/datastructure/market"
	"github.com/alessiosavi/GoArbitrage/utils"
	fileutils "github.com/alessiosavi/GoGPUtils/files"

	req "github.com/alessiosavi/Requests"
)

type Kraken struct {
	PairsNames []string
	Pairs      map[string]datastructure.KrakenPair
	OrderBook  map[string]datastructure.BitfinexOrderBook
}

const KRAKEN_PAIRS_URL string = `https://api.bitfinex.com/v1/symbols`
const KRAKEN_PAIRS_DETAILS_URL string = `https://api.kraken.com/0/public/AssetPairs`
const KRAKEN_ORDER_BOOK_URL string = `https://api.kraken.com/0/public/Depth?pair=`

var KRAKEN_PAIRS_DATA string = path.Join(constants.KRAKEN_PATH, "pairs_list.json")
var KRAKEN_PAIRS_DETAILS string = path.Join(constants.KRAKEN_PATH, "pairs_info.json")
var KRAKEN_ORDERBOOK_DATA string = path.Join(constants.KRAKEN_PATH, "orders/")

func (k *Kraken) Init() {
	k.Pairs = make(map[string]datastructure.KrakenPair)
	k.OrderBook = make(map[string]datastructure.BitfinexOrderBook)
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

	} else {
		zap.S().Debugw("Sendind request to [" + KRAKEN_PAIRS_DETAILS_URL + "]")
		// Call the HTTP method for retrieve the pairs
		resp := request.SendRequest(KRAKEN_PAIRS_DETAILS_URL, "GET", nil, false)
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
		var order datastructure.BitfinexOrderBook
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
				log.Println("Order: ", order)
			}
		} else {
			url := KRAKEN_ORDER_BOOK_URL + pair + `&count=4`
			zap.S().Debugw("Sendind request to [" + url + "]")
			// Call the HTTP method for retrieve the pairs
			resp := request.SendRequest(url, "GET", nil, false)
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

func loadOrderBook(data []byte) (datastructure.BitfinexOrderBook, error) {
	res := &datastructure.Response{}
	var err error
	if err = json.Unmarshal(data, res); err != nil {
		zap.S().Warn("ERROR! :" + err.Error())
		return datastructure.BitfinexOrderBook{}, err
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
	return datastructure.BitfinexOrderBook{}, err
}

// loadKrakenPairs is delegated to load the data related to pairs info
func loadKrakenPairs(data []byte) map[string]datastructure.KrakenPair {
	var err error

	// Creating the maps for the JSON data
	m := map[string]interface{}{}
	var pairInfo map[string]datastructure.KrakenPair = make(map[string]datastructure.KrakenPair)

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

func (k *Kraken) GetMarketData(pair string) (market.Market, error) {
	var markets market.Market
	markets.Asks = make(map[string][]market.MarketOrder, len(k.OrderBook))
	markets.Bids = make(map[string][]market.MarketOrder, len(k.OrderBook))
	markets.MarketName = `GEMINI`
	var order market.MarketOrder
	if orders, ok := k.OrderBook[pair]; ok {
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
