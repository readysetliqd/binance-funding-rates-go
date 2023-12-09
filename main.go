package main

import (
	"log"
)

func main() {
	// Define stablecoins
	stableCoins := []string{"USDC", "USDT", "DAI", "VAI", "RAI", "UST", "USDD", "TUSD", "LUSD", "XUSD"}
	// Load Environment variables
	// Connect to database
	// Create table "topN funding rates" primary keys fundingTime and symbol
	// For snapshot_date in table: Query marketcap_snapshots table for Top N non stablecoins, make slice
	// For tickers in slice: poll binance public api for fundingrate info
	// ex. url https://fapi.binance.com/fapi/v1/fundingRate?symbol=XRPUSDT&startTime=1568102400000&endTime=1578297500000&limit=1000
	// unmarshal api JSON to struct
	// if binance public api doesn't have funding rate history for coin, try next ticker from marketcap_snapshots table
	// insert data to "topN funding rates" table
	// after "topN funding rates" table caught up to last entry of marketcap_snapshots
	// create new "funding averages" table foreign key fundingTime
	// calculate and insert average and median funding rates for fundingTime
	// calculate marketcap weighted ex-BTC average and median funding rates
	// calculate rolling 1 month 3 month 1 year funding rate confidence interval
	// insert all data to new "funding averages" table
	// after "funding averages" table caught up to last entry
	// query to find all fundingTime where average funding rates are at extremes

	log.Println("Hello world")
}
