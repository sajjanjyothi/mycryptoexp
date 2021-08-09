package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sajjanjyothi/bitstamp"
)

const (
	MIN_CRYPTO_PRICE = 15.0
	MAX_CRYPTO_PRICE = 23.0
	BUYING_AMOUNT    = 25
	CURRENCY         = "linkgbp"
)

type Subscribe struct {
	Event string      `json:"event"`
	Data  interface{} `json:"data"`
}

type LiveOrder struct {
	Data struct {
		Price_str string `json:"price_str"`
	} `json:"data"`
}

var orderInHand int //to check whether we have an order already in place
var lastBuyPrice float64
var isloss bool
var lastSoldPrice float64

func main() {
	bistampClient := bitstamp.BitStamp{}
	fmt.Println(bistampClient.GetData("/api/v2/balance/", nil))

	c, res, err := websocket.DefaultDialer.Dial("wss://ws.bitstamp.net/", nil)
	if err != nil {
		panic(err)
	}

	subscription := Subscribe{
		Event: "bts:subscribe",
		Data: struct {
			Channel string `json:"channel"`
		}{
			Channel: "live_trades_" + CURRENCY,
		},
	}
	data, err := json.Marshal(subscription)
	if err != nil {
		panic(err)
	}
	fmt.Println("JSON data " + string(data))
	err = c.WriteMessage(websocket.TextMessage, data)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				panic(err)
			}

			if !strings.Contains(string(message), "price_str") {
				continue
			}

			price := &LiveOrder{}
			err = json.Unmarshal(message, &price)
			if err != nil {
				panic(err)
			}
			log.Printf("recv: %s", message)
			log.Println(price.Data.Price_str)
			currencyValue, _ := strconv.ParseFloat(price.Data.Price_str, 10)
			log.Println(currencyValue)

			if currencyValue < MIN_CRYPTO_PRICE {
				if orderInHand == 0 {
					if !isloss || currencyValue > lastSoldPrice { //not a loss and seems like an increase happening
						log.Println("Buying with currency value " + price.Data.Price_str)
						bistampClient := bitstamp.BitStamp{}
						values := url.Values{}
						values.Add("amount", strconv.Itoa(BUYING_AMOUNT))
						res, err := bistampClient.GetData("/api/v2/buy/instant/"+CURRENCY+"/", values)
						if err != nil {
							log.Println(err)
							continue
						}
						log.Println(res)
						lastBuyPrice = currencyValue
						orderInHand = 1
						isloss = false
					}

				} else { //already have an order- price going down - sell it
					if currencyValue < lastBuyPrice { //current price is less than last buy price sell it
						log.Println("selling with currency value " + price.Data.Price_str)
						bistampClient := bitstamp.BitStamp{}
						values := url.Values{}
						values.Add("amount", strconv.Itoa(BUYING_AMOUNT))
						res, err := bistampClient.GetData("/api/v2/sell/instant/"+CURRENCY+"/", values)
						if err != nil {
							log.Println(err)
							continue
						}
						log.Println(res)
						orderInHand = 0
						lastSoldPrice = currencyValue
						isloss = true
					}
				}
			} else if currencyValue >= MAX_CRYPTO_PRICE {
				if orderInHand == 1 { //profit
					log.Println("selling with currency value " + price.Data.Price_str)
					bistampClient := bitstamp.BitStamp{}
					values := url.Values{}
					values.Add("amount", strconv.Itoa(BUYING_AMOUNT))
					res, err := bistampClient.GetData("/api/v2/sell/instant/"+CURRENCY+"/", values)
					if err != nil {
						log.Println(err)
						continue
					}
					log.Println(res)
					orderInHand = 0
					isloss = false
				}
			} else {
				log.Println("Not doing anything- currency price ", currencyValue)
			}

		}
	}()
	for {
		time.Sleep(10 * time.Second)
	}
	fmt.Println(res)
	defer c.Close()
}
