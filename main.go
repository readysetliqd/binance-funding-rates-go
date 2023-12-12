package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/readysetliqd/binance-funding-rates-go/data"
)

// Table name of already existing table with weekly snapshots of all cryptocurrencies ranked
// by marketca. Table built from github.com/readysetliqd/cryto-historical-marketcas-scraper-go
const snapshotsTableName = "marketcap_snapshots"

// Table name in database will be affected by this const. Number of cryptos to
// analyze from every snapshot ie. top 10 by market cap
const topN = 100

// First snapshot entry from CoinMarketCap is April 28th, 2013
// var dataStartDate = time.Date(2013, 4, 28, 0, 0, 0, 0, time.UTC)
// Binance futures went live Sep. 13, 2019. First CoinMarketCap snapshot entry after that was the 15th
var dataStartDate = time.Date(2019, 9, 15, 0, 0, 0, 0, time.UTC)

var fundingTableName = "top" + strconv.Itoa(topN) + "_historical_funding_rates"

func main() {
	// #region Load Environment variables
	err := godotenv.Load("db.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// #endregion

	// #region Connect to database
	ctx := context.Background()
	connStr := "postgres://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASS") + "@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("DB_NAME")
	dbpool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		log.Printf("Unable to connect to database: %v\n", os.Stderr)
		os.Exit(1)
	} else {
		log.Println("DB connected successfully")
	}
	defer dbpool.Close()
	// #endregion

	// #region Create table "topN_funding_rates" if not exists; initialize date
	var date time.Time
	var tableExists bool
	queryTableExists := `SELECT EXISTS (
		SELECT 1
		FROM information_schema.tables
		WHERE table_name = '` + fundingTableName + `'
		) AS table_existence;`
	err = dbpool.QueryRow(ctx, queryTableExists).Scan(&tableExists)
	if err != nil {
		log.Fatal("QueryRow failed | ", err)

	}

	if !tableExists {
		log.Printf("Table %s does not exist. Creating table...", fundingTableName)
		queryCreateTable := `CREATE TABLE ` + fundingTableName + `(
			snapshot_date DATE,
			symbol TEXT,
			funding_time BIGINT NOT NULL,
			funding_rate DECIMAL NOT NULL,
			mark_price DECIMAL,
			
			PRIMARY KEY (symbol, funding_time)
			);
			`
		_, err = dbpool.Exec(ctx, queryCreateTable)
		if err != nil {
			log.Fatalf("Unable to create the '%s' table | %v", fundingTableName, err)
		}
		date = dataStartDate
	} else {
		log.Printf("Table %s exists. Fetching last entry", fundingTableName)
		queryLastDate := dbpool.QueryRow(ctx, `SELECT snapshot_date FROM `+fundingTableName+`ORDER BY snapshot_date DESC LIMIT 1`)
		queryLastDate.Scan(&date)
		if date.Before(dataStartDate) { // fixes date when table exists but no entries
			date = dataStartDate
		} else {
			date = date.AddDate(0, 0, 7) // if entries exists, sets date to next weekly snapshot
		}
		log.Println(date)
	}
	// #endregion

	// #region Make a snapshots slice for snapshot_dates without entries in snapshotsTableName table
	var snapshots []time.Time
	snapshotRows, err := dbpool.Query(ctx, `SELECT snapshot_date FROM `+snapshotsTableName+` WHERE snapshot_date >= '`+date.Format("2006-01-02")+`' GROUP BY snapshot_date ORDER BY snapshot_date ASC`)
	if err != nil {
		log.Fatal("error querying rows | ", err)
	}
	snapshots, err = pgx.CollectRows(snapshotRows, pgx.RowTo[time.Time])
	if err != nil {
		log.Fatal("error scanning rows | ", err)
	}
	log.Println(snapshots)
	// #endregion

	// #region Check for restricted location
	var msg []byte
	resp := map[string]interface{}{}
	url := "https://fapi.binance.com/fapi/v1/fundingRate"
	res, err := http.Get(url)
	if err != nil {
		log.Fatal("http.Get error | ", err)
	}
	msg, err = io.ReadAll(res.Body)
	if err != nil {
		log.Fatal("io.ReadAll error | ", err)
	}
	json.Unmarshal(msg, &resp)
	if resp["msg"] != nil {
		if strings.Contains(resp["msg"].(string), "restricted location") {
			log.Fatal("IP is being geoblocked, check location or VPN | ", resp["msg"])
		}
	}
	// #endregion

	for _, snapshot := range snapshots {
		// #region Set slice of symbols to check for funding rate history on Binance
		var symbols []string
		// HARDCODED WORKAROUND
		// Binance futures only had 3 markets in 2019, 80 markets in 2020.
		// Hardcoding the symbols list speeds up the data collection by avoiding
		// unneccessary API calls especially with higher values for topN const (20+)
		// Symbols lists made by hand reading through Binance announcements
		if snapshot.Before(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)) {
			symbols = data.SymbolsBefore2020
		} else if snapshot.Before(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)) {
			symbols = data.SymbolsBefore2021
		} else if snapshot.Before(time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)) {
			symbols = data.SymbolsBefore2022
		} else if snapshot.Before(time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)) {
			symbols = data.SymbolsBefore2023
		} else {
			var stableCoinsQuery = strings.Join(data.StableCoins, ",")
			symbolRows, err := dbpool.Query(ctx, `SELECT symbol FROM `+snapshotsTableName+` WHERE snapshot_date = '`+snapshot.Format("2006-01-02")+`' AND symbol NOT IN (`+stableCoinsQuery+`) GROUP BY symbol, rank ORDER BY rank ASC`)
			if err != nil {
				log.Fatal("error sending query", err)
			}
			symbols, err = pgx.CollectRows(symbolRows, pgx.RowTo[string])
			if err != nil {
				log.Fatal("error collecting rows", err)
			}
			// adjust for binance specific perps listings
			for i, symbol := range symbols {
				if symbol == "LUNC" || symbol == "SHIB" || symbol == "PEPE" ||
					symbol == "FLOKI" || symbol == "SATS" || symbol == "BONK" {
					symbols[i] = "1000" + symbol
				}
			}
		}
		// #endregion

		var queuedApiResp []data.FundingRateApiResp
		countCoinsApiResp := 0
		for _, symbol := range symbols {
			url = fmt.Sprintf("https://fapi.binance.com/fapi/v1/fundingRate?symbol=%sUSDT&startTime=%v&endTime=%v", symbol, snapshot.UnixMilli(), snapshot.AddDate(0, 0, 7).UnixMilli()-1)
			log.Println(url)
			res, err := http.Get(url)
			if err != nil {
				log.Println("http.Get error | ", err)
			}
			defer res.Body.Close()
			msg, err = io.ReadAll(res.Body)
			if err != nil {
				log.Fatal("io.ReadAll error | ", err)
			}
			var fundingRates []data.FundingRateApiResp
			json.Unmarshal(msg, &fundingRates)
			log.Println("Number of funding rates: ", len(fundingRates))
			if len(fundingRates) < 21 {
				log.Println("Not enough funding rate history in this period. Skipping coin | ", symbol)
			} else {

				countCoinsApiResp += 1
				queuedApiResp = append(queuedApiResp, fundingRates...)
			}
			// Break loop if topN number of coins queued
			if countCoinsApiResp >= topN {
				break
			}
			time.Sleep(500 * time.Millisecond)
		}
		// TODO
		// loop for apiResp in queuedApiResp convert data types for Row struct
		// and append to queuedRows. Add snapshot date, rank, marketcap w/e else
		// other table. Ask GPT what some smart relations are
		// still need to make the Row struct in data package and initialize queuedRows
	}
	// insert data to "topN funding rates" table
	// after "topN funding rates" table caught up to last entry of marketcap_snapshots
	// create new "funding averages" table foreign key fundingTime
	// calculate and insert average and median funding rates for fundingTime
	// calculate marketcap weighted ex-BTC average and median funding rates
	// calculate rolling 1 month 3 month 1 year funding rate confidence interval
	// insert all data to new "funding averages" table
	// after "funding averages" table caught up to last entry
	// query to find all fundingTime where average funding rates are at extremes
}
