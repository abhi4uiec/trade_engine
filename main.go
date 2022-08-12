package main

import (
	"binance/engine"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	_ "github.com/mattn/go-sqlite3"
	"github.com/shopspring/decimal"
)

var sqliteDatabase *sql.DB

var done bool = false

type Trade struct {
	UpdateId int64  `json:"u"`
	Symbol   string `json:"s"`
	BidPrice string `json:"b"`
	BidQty   string `json:"B"`
	AskPrice string `json:"a"`
	AskQty   string `json:"A"`
}

type request struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int      `json:"id"`
}

func main() {

	symbol, tradeType, quantity, price := acceptInput()

	initializeDB()

	// websocket source
	c, _, err := websocket.DefaultDialer.Dial("wss://stream.binance.com/ws/"+symbol+"@bookticker", nil)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	if err != nil {
		panic(err)
	}
	defer c.Close()

	// create the input channel
	inputStocks := make(chan Trade)

	// subscribe to binance service
	markPriceReq := request{"SUBSCRIBE", []string{symbol + "@bookTicker"}, 1}
	c.WriteJSON(markPriceReq)

	// create the order book
	book := engine.OrderBook{
		Bids: make([]engine.Order, 0, 100),
		Asks: make([]engine.Order, 0, 100),
	}

	var side engine.Side
	var qty, amount decimal.Decimal

	if strings.EqualFold(tradeType, "buy") {
		side = 1
	} else if strings.EqualFold(tradeType, "sell") {
		side = 0
	}

	if p, _ := strconv.Atoi(price); p > 0 {
		qty = decimalValue(quantity)
		amount = decimalValue(price)
	}
	order := engine.Order{ID: int64(rand.Intn(64)), Side: side, Quantity: qty, Price: amount, Timestamp: int64(rand.Intn(64))}

	//dones := make(chan struct{})

	// producer: read from websocket and send to channel
	go func() {
		// read from the websocket
		for {

			if done {
				// DISPLAY INSERTED RECORDS
				displayRecords(sqliteDatabase)
				fmt.Println("All orders processed")
				break
			}

			_, message, err := c.ReadMessage()
			if err != nil {
				break
			}
			// unmarshal the message
			var trade Trade
			json.Unmarshal(message, &trade)
			// send the trade to the channel
			inputStocks <- trade

		}
		close(inputStocks)
	}()

	// print the trades
	for trade := range inputStocks {
		json, _ := json.Marshal(trade)
		fmt.Println(string(json))

		bidPrice, _ := strconv.ParseFloat(trade.BidPrice, 64)

		if bidPrice > 0 {

			book.AddBuyOrder(engine.Order{ID: trade.UpdateId, Side: 1, Quantity: decimalValue(trade.BidQty), Price: decimalValue(trade.BidPrice), Timestamp: 0})

			book.AddSellOrder(engine.Order{ID: trade.UpdateId, Side: 0, Quantity: decimalValue(trade.AskQty), Price: decimalValue(trade.AskPrice), Timestamp: 0})

			// process the order
			trades := book.Process(order)

			log.Println("Trades length: ", len(trades))

			if len(trades) != 0 {
				for _, trade := range trades {
					liteDB(trade)
				}
			}

			//Check if all orders were processed
			if order.Side == 0 {
				done = true
				n := len(book.Asks)
				for i := 0; i < n; i++ {
					if order.ID == book.Asks[i].ID {
						order = book.Asks[i]
						done = false
						break
					}
				}
			} else if order.Side == 1 {
				done = true
				n := len(book.Bids)
				for i := 0; i < n; i++ {
					if order.ID == book.Bids[i].ID {
						order = book.Bids[i]
						done = false
						break
					}
				}
			}
		}
	}
}

func acceptInput() (string, string, string, string) {
	var symbol, tradeType, quantity, price string

	fmt.Print("Enter your Trading Pair: ")
	_, err := fmt.Scan(&symbol)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Trading Pair", symbol)

	fmt.Print("Enter your TradeType(Sell/Buy): ")
	_, err = fmt.Scan(&tradeType)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("TradeType", tradeType)

	fmt.Print("Enter your Quantity: ")
	_, err = fmt.Scan(&quantity)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Quantity", quantity)

	fmt.Print("Enter your price: ")
	_, err = fmt.Scan(&price)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("price", price)
	return symbol, tradeType, quantity, price
}

func initializeDB() {
	os.Remove("sqlite-database.db") // I delete the file to avoid duplicated records.
	// SQLite is a file based database.

	log.Println("Creating sqlite-database.db...")
	file, err := os.Create("sqlite-database.db") // Create SQLite file
	if err != nil {
		log.Fatal(err.Error())
	}
	file.Close()
	log.Println("sqlite-database.db created")

	sqliteDatabase, _ = sql.Open("sqlite3", "./sqlite-database.db") // Open the created SQLite File
	createTable(sqliteDatabase)                                     // Create Database Tables
}

func liteDB(trades engine.Trade) {

	// INSERT RECORDS
	insertRecords(sqliteDatabase, trades.TakerOrderID, trades.MakerOrderID, trades.Quantity, trades.Price)
}

func createTable(db *sql.DB) {
	createTableSQL := `CREATE TABLE sell_orders (
		"id" integer NOT NULL PRIMARY KEY AUTOINCREMENT,		
		"taker_order_id" integer,
		"maker_order_id" integer,
		"quantity" NUMERIC,
		"price" NUMERIC	
	  );` // SQL Statement for Create Table

	log.Println("Create sell_order table...")
	statement, err := db.Prepare(createTableSQL) // Prepare SQL Statement
	if err != nil {
		log.Fatal(err.Error())
	}
	statement.Exec() // Execute SQL Statements
	log.Println("sell_orders table created")
}

// We are passing db reference connection from main to our method with other parameters
func insertRecords(db *sql.DB, taker_order_id int64, maker_order_id int64, quantity decimal.Decimal, price decimal.Decimal) {
	log.Println("Inserting record ...")
	insertSQL := `INSERT INTO sell_orders(taker_order_id, maker_order_id, quantity, price) VALUES (?, ?, ?, ?)`
	statement, err := db.Prepare(insertSQL) // Prepare statement.
	// This is good to avoid SQL injections
	if err != nil {
		log.Fatalln(err.Error())
	}
	_, err = statement.Exec(taker_order_id, maker_order_id, quantity, price)
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func displayRecords(db *sql.DB) {
	defer db.Close()
	row, err := db.Query("SELECT * FROM sell_orders order BY id")
	if err != nil {
		log.Fatal(err)
	}
	defer row.Close()
	for row.Next() { // Iterate and fetch the records from result cursor
		var id int
		var taker_order_id int64
		var maker_order_id int64
		var quantity decimal.Decimal
		var price decimal.Decimal
		row.Scan(&id, &taker_order_id, &maker_order_id, &quantity, &price)
		log.Println("Sell order records: ", "TakerOrderId=", taker_order_id, " MakerOrderId", maker_order_id, " Quantity", quantity, " Price", price)
	}
}

func decimalValue(str string) decimal.Decimal {
	v, _ := decimal.NewFromString(str)
	return v
}
