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

	constants "github.com/alessiosavi/GoArbitrage/datastructure"
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
	PairsName []string                        `json:"pairs_name"`
	Pairs     []datastructure.OkCoinPairs     `json:"pairs"`
	OrderBook []datastructure.OkCoinOrderBook `json:"orderbook"`
	MakerFee  float32                         `json:"maker_fee"`
	TakerFees float32                         `json:"taker_fee"`
	// FeePercent is delegated to save if the fee is in percent or in coin
	FeePercent bool `json:"fee_percent"`
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

	for i := range pairsInfo {
		pairsInfo[i].Pair = pairsInfo[i].BaseCurrency + "-" + pairsInfo[i].QuoteCurrency
	}

	ok.Pairs = pairsInfo
	// Update the file with the new data
	utils.DumpStruct(pairsInfo, OKCOIN_PAIRS_DETAILS)
	return nil
}

func (ok *OkCoin) GetAllOrderBook() error {
	var request req.Request
	var orders []datastructure.OkCoinOrderBook
	var data []byte
	var err error

	for _, pair := range ok.PairsName {
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
				return err
			}
		} else {
			time.Sleep(100 * time.Millisecond)
			url := OKCOIN_PAIRS_DETAILS_URL + pair + "/book"
			zap.S().Debugw("Sendind request to [" + url + "]")
			// Call the HTTP method for retrieve the pairs
			resp := request.SendRequest(url, "GET", nil, false)
			if resp.Error != nil {
				zap.S().Debugw("Error during http request. Err: " + resp.Error.Error())
				return resp.Error
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
		}
		orders = append(orders, orderbook)
		// Update the file with the new data
		utils.DumpStruct(orderbook, file_data)
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
