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

type BitfinexOrderBook struct {
	Pair string          `json:"pair"`
	Asks []BitfinexOrder `json:"asks"`
	Bids []BitfinexOrder `json:"bids"`
}

type BitfinexOrder struct {
	Price     string
	Volume    string
	Timestamp time.Time
}

type Response struct {
	Error  []interface{}                `json:"error"`
	Result map[string]BitfinexOrderBook `json:"result"`
}

// UnmarshalJSON decode a BifinexOrder.
func (b *BitfinexOrder) UnmarshalJSON(data []byte) error {
	var packedData []json.Number
	err := json.Unmarshal(data, &packedData)
	if err != nil {
		return err
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
