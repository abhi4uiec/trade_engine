package engine

import (
	"encoding/json"

	"github.com/shopspring/decimal"
)

type Trade struct {
	TakerOrderID int64           `json:"taker_order_id"`
	MakerOrderID int64           `json:"maker_order_id"`
	Quantity     decimal.Decimal `json:"quantity"`
	Price        decimal.Decimal `json:"price"`
	Timestamp    int64           `json:"timestamp"`
}

// struct to json
func (trade *Trade) FromJSON(msg []byte) error {
	return json.Unmarshal(msg, trade)
}

func (trade *Trade) ToJSON() []byte {
	str, _ := json.Marshal(trade)
	return str
}
