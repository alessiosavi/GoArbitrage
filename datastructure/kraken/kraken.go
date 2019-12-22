package kraken

import (
	"encoding/json"
	"time"
)

type KrakenPair struct {
	Altname           string      `json:"altname"`
	Base              string      `json:"base"`
	Quote             string      `json:"quote"`
	PairDecimals      int         `json:"pair_decimals"`
	LotDecimals       int         `json:"lot_decimals"`
	LotMultiplier     int         `json:"lot_multiplier"`
	Fees              [][]float64 `json:"fees"`
	FeesMaker         [][]float64 `json:"fees_maker"`
	FeeVolumeCurrency string      `json:"fee_volume_currency"`
}

type KrakenOrderBook struct {
	Pair string        `json:"pair"`
	Asks []KrakenOrder `json:"asks"`
	Bids []KrakenOrder `json:"bids"`
}

type KrakenOrder struct {
	Price     string    `json:"price"`
	Volume    string    `json:"volume"`
	Timestamp time.Time `json:"timestamp"`
}
type KrakenOrderJson KrakenOrder

type Response struct {
	Error  []interface{}              `json:"error"`
	Result map[string]KrakenOrderBook `json:"result"`
}

// UnmarshalJSON decode a BifinexOrder.
func (b *KrakenOrder) UnmarshalJSON(data []byte) error {
	var packedData []json.Number
	var err error
	err = json.Unmarshal(data, &packedData)
	if err != nil {
		var order KrakenOrderJson
		err = json.Unmarshal(data, &order)
		if err != nil {
			return err
		}
		b.Price = order.Price
		b.Volume = order.Volume
		b.Timestamp = order.Timestamp
		return nil
	}
	b.Price = packedData[0].String()
	b.Volume = packedData[1].String()
	t, err := packedData[2].Int64()
	if err != nil {
		return err
	}
	b.Timestamp = time.Unix(t, 0)
	return nil
}
