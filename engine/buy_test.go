package engine

import (
	"testing"

	"github.com/shopspring/decimal"
)

func TestBuyLimitProcess(t *testing.T) {

	// create the order book
	book := OrderBook{
		Bids: make([]Order, 0, 10),
		Asks: make([]Order, 0, 10),
	}

	order := Order{ID: 10, Side: 1, Quantity: decimalValue("50"), Price: decimalValue("100"), Timestamp: 0}

	book.AddSellOrder(Order{ID: 20, Side: 0, Quantity: decimalValue("100"), Price: decimalValue("60"), Timestamp: 0})

	trades := book.Process(order)

	if len(trades) == 0 {
		t.Fatal("OrderBook failed to process buy limit order (done is not empty)")
	}

	t.Log("TestBuyLimitProcess test case passed")

}

func TestSellLimitProcess(t *testing.T) {

	// create the order book
	book := OrderBook{
		Bids: make([]Order, 0, 10),
		Asks: make([]Order, 0, 10),
	}

	order := Order{ID: 10, Side: 0, Quantity: decimalValue("50"), Price: decimalValue("100"), Timestamp: 0}

	book.AddBuyOrder(Order{ID: 20, Side: 1, Quantity: decimalValue("100"), Price: decimalValue("120"), Timestamp: 0})

	trades := book.Process(order)

	if len(trades) == 0 {
		t.Fatal("OrderBook failed to process sell limit order (done is not empty)")
	}

	t.Log("TestSellLimitProcess test case passed")

}

func decimalValue(str string) decimal.Decimal {
	v, _ := decimal.NewFromString(str)
	return v
}
