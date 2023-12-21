package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
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

// Table name in database that will be created by this program and filled with
// historical funding rate data
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
			funding_time BIGINT NOT NULL,
			symbol TEXT,
			funding_rate DECIMAL NOT NULL,
			mark_price DECIMAL,
			snapshot_date DATE,
			rank INTEGER,

			PRIMARY KEY (symbol, funding_time),
			FOREIGN KEY (snapshot_date, rank, symbol) REFERENCES ` + snapshotsTableName + `(snapshot_date, rank, symbol)
			);
			`
		_, err = dbpool.Exec(ctx, queryCreateTable)
		if err != nil {
			log.Fatalf("Unable to create the '%s' table | %v", fundingTableName, err)
		}
		date = dataStartDate
	} else {
		log.Printf("Table %s exists. Fetching last entry", fundingTableName)
		queryLastDate := dbpool.QueryRow(ctx, `SELECT snapshot_date FROM `+fundingTableName+` ORDER BY snapshot_date DESC LIMIT 1`)
		queryLastDate.Scan(&date)
		log.Println(date)
		if date.Before(dataStartDate) { // fixes date when table exists but no entries
			date = dataStartDate
		} else {
			date = date.AddDate(0, 0, 7) // if entries exists, sets date to next weekly snapshot
		}
		log.Println("Starting queries at date: ", date)
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
		log.Fatal("error collecting rows | ", err)
	}
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

	// Iterate over slice of snapshots that have yet to be added to database
	for _, snapshot := range snapshots {
		// #region Set slice of symbols to check for funding rate history on Binance
		var symbols []string
		// HARDCODED WORKAROUND
		// Binance futures only had 3 markets in 2019, 80 markets in 2020.
		// Hardcoding the symbols list speeds up the data collection by avoiding
		// unneccessary API calls especially with higher values for topN const (20+)
		// Symbols lists made by hand reading through Binance announcements
		switch {
		case snapshot.Before(time.Date(2020, time.January, 1, 0, 0, 0, 0, time.UTC)):
			symbols = data.SymbolsBefore2020
		case snapshot.Before(time.Date(2021, time.January, 1, 0, 0, 0, 0, time.UTC)):
			symbols = data.SymbolsBefore2021
		case snapshot.Before(time.Date(2022, time.January, 1, 0, 0, 0, 0, time.UTC)):
			symbols = data.SymbolsBefore2022
		case snapshot.Before(time.Date(2023, time.January, 1, 0, 0, 0, 0, time.UTC)):
			symbols = data.SymbolsBefore2023
		case snapshot.Before(time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)):
			symbols = data.SymbolsBefore2024
		default:
			var stableCoinsQuery = strings.Join(data.StableCoins, ",")
			symbolRows, err := dbpool.Query(ctx, `SELECT symbol FROM `+snapshotsTableName+` WHERE snapshot_date = '`+snapshot.Format("2006-01-02")+`' AND symbol NOT IN (`+stableCoinsQuery+`) GROUP BY symbol, rank ORDER BY rank ASC`)
			if err != nil {
				log.Fatal("error sending query", err)
			}
			symbols, err = pgx.CollectRows(symbolRows, pgx.RowTo[string])
			if err != nil {
				log.Fatal("error collecting rows", err)
			}

			// adjust for binance specific perps listings and remove CMC duplicates
			seen := make(map[string]bool)
			symbolsNoDuplicates := []string{}
			for i, symbol := range symbols {
				for _, thousandSymbol := range data.ThousandSymbols {
					if symbol == thousandSymbol {
						symbols[i] = "1000" + symbol
					}
				}
				if _, ok := seen[symbols[i]]; !ok {
					seen[symbols[i]] = true
					symbolsNoDuplicates = append(symbolsNoDuplicates, symbols[i])
				}
			}
			symbols = symbolsNoDuplicates
		}
		// #endregion

		// #region Build slice of symbols with their ranks pulled from database
		var symbolStructs []data.Symbol
		for _, symbol := range symbols {
			if strings.Contains(symbol, "1000") {
				_, symbol, _ = strings.Cut(symbol, "1000")
			}
			var rank int64
			rankRow := dbpool.QueryRow(ctx, `SELECT rank FROM `+snapshotsTableName+` WHERE snapshot_date = '`+snapshot.Format("2006-01-02")+`' AND symbol = '`+symbol+`'`)
			err = rankRow.Scan(&rank)
			if err != nil {
				if strings.Contains(err.Error(), "no rows in result set") {
					// Handle edge case where IOTA on binance is MIOTA on CMC
					if symbol == "IOTA" {
						symbol = "MIOTA"
						rankRow := dbpool.QueryRow(ctx, `SELECT rank FROM `+snapshotsTableName+` WHERE snapshot_date = '`+snapshot.Format("2006-01-02")+`' AND symbol = '`+symbol+`'`)
						err = rankRow.Scan(&rank)
						if err != nil {
							if strings.Contains(err.Error(), "no rows in result set") {
								continue
							} else {
								log.Fatal("Error scanning row | ", err, symbol)
							}
						}
					} else {
						continue
					}
				} else {
					log.Fatal("Error scanning row | ", err, symbol)
				}
			}
			var newSymbol = data.Symbol{
				Symbol: symbol,
				Rank:   rank,
			}
			if newSymbol.Rank != 0 {
				symbolStructs = append(symbolStructs, newSymbol)
			}
		}
		sort.Slice(symbolStructs[:], func(i, j int) bool {
			return symbolStructs[i].Rank < symbolStructs[j].Rank
		}) // #endregion

		// #region Iterate over symbols and poll binance fundingRate API. Add...
		// funding history for coin if data is complete between this snapshot
		// and the next until list is exhausted or topN coins with complete data
		// is reached, whichever comes first
		var queuedApiResp []data.FundingRateApiResp
		countCoinsApiResp := 0
		for _, symbol := range symbolStructs {
			url = fmt.Sprintf("https://fapi.binance.com/fapi/v1/fundingRate?symbol=%sUSDT&startTime=%v&endTime=%v", symbol.Symbol, snapshot.UnixMilli(), snapshot.AddDate(0, 0, 7).UnixMilli()-1)
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
			if len(fundingRates) < 21 {
				continue
			} else {
				countCoinsApiResp += 1
				queuedApiResp = append(queuedApiResp, fundingRates...)
			}
			// Break loop if topN number of coins queued
			if countCoinsApiResp >= topN {
				break
			}
			// binance funding rate history rate limit: 500/5min/IP
			time.Sleep(600 * time.Millisecond)
		} // #endregion

		// #region Iterate over APIresps and build slice of rows to batch insert to db
		var queuedRows []data.Row
		for _, apiResp := range queuedApiResp {
			rate, err := strconv.ParseFloat(apiResp.Rate, 64)
			var mark sql.NullFloat64
			if apiResp.Mark == "" {
				mark.Float64 = 0.0
				mark.Valid = false
			} else {
				mark.Float64, err = strconv.ParseFloat(apiResp.Mark, 64)
				mark.Valid = true
			}
			if err != nil {
				log.Fatal("ParseFloat error | ", err)
			}

			symbol, _, ok := strings.Cut(apiResp.Symbol, "USDT")
			if !ok {
				log.Fatal("Error cutting USDT from string | ", err)
			}
			if strings.Contains(symbol, "1000") {
				_, symbol, _ = strings.Cut(symbol, "1000")
			}

			var rank int64
			found := false
			for _, symbolStruct := range symbolStructs {
				if symbol == symbolStruct.Symbol {
					rank = symbolStruct.Rank
					found = true
					break
				}
			}
			if !found {
				if symbol == "IOTA" {
					symbol = "MIOTA"
					for _, symbolStruct := range symbolStructs {
						if symbol == symbolStruct.Symbol {
							rank = symbolStruct.Rank
							break
						}
					}
				} else {
					log.Fatal("Symbol not found in slice | ", symbol)
				}
			}
			newRow := data.Row{
				FundingTime:  apiResp.Time,
				Symbol:       symbol,
				FundingRate:  rate,
				MarkPrice:    mark,
				SnapshotDate: snapshot,
				Rank:         rank,
			}
			queuedRows = append(queuedRows, newRow)
		} // #endregion

		// #region Batch insert queuedRows to database
		queryInsertData := `
			INSERT INTO ` + fundingTableName + `
			(funding_time, symbol, funding_rate, mark_price, snapshot_date, rank)
			VALUES ($1, $2, $3, $4, $5, $6);
			`
		batch := &pgx.Batch{}
		for _, row := range queuedRows {
			batch.Queue(queryInsertData, row.FundingTime, row.Symbol, row.FundingRate, row.MarkPrice, row.SnapshotDate, row.Rank)
		}
		br := dbpool.SendBatch(ctx, batch)
		_, err := br.Exec()
		if err != nil {
			log.Fatal("Unable to execute statement in batch queue | ", err)
		}
		log.Printf("Successfully inserted %d rows to table %s at snapshot_date %s", len(queuedRows), fundingTableName, snapshot)

		err = br.Close()
		if err != nil {
			log.Fatal("Error closing batch | ", err)

		} // #endregion
	}
	log.Printf("Insertions to table %s have caught up to entries in table %s", fundingTableName, snapshotsTableName)

	// #region Build list of snapshot_dates with incomplete mark price data
	snapshotRows, err = dbpool.Query(ctx, `SELECT snapshot_date FROM `+fundingTableName+` WHERE mark_price IS NULL GROUP BY snapshot_date ORDER BY snapshot_date ASC`)
	if err != nil {
		log.Fatal("error querying rows | ", err)
	}
	snapshots, err = pgx.CollectRows(snapshotRows, pgx.RowTo[time.Time])
	if err != nil {
		log.Fatal("error collecting rows | ", err)
	} // #endregion

	// Iterate over snapshots and find symbols without mark_price data
	for _, snapshot := range snapshots {
		// #region Build list of symbols without mark_price data at snapshot_date
		var symbols []string
		symbolRows, err := dbpool.Query(ctx, `SELECT symbol FROM `+fundingTableName+`WHERE snapshot_date = `+snapshot.Format("2006-01-02")+` AND mark_price IS NULL GROUP BY symbol`)
		if err != nil {
			log.Fatal("error querying rows | ", err)
		}
		symbols, err = pgx.CollectRows(symbolRows, pgx.RowTo[string])
		if err != nil {
			log.Fatal("error collecting rows | ", err)
		} // #endregion

		// Iterate over list of symbols and fill in mark_price data from api
		for _, symbol := range symbols {
			// #region build api url and poll it with http.Get
			if symbol == "MIOTA" {
				symbol = "IOTA"
			}
			symbol = symbol + "USDT"
			url = fmt.Sprintf("https://fapi.binance.com/fapi/v1/markPriceKlines?symbol=%s&interval=8h&limit=21&startTime=%v&endTime=%v", symbol, snapshot.UnixMilli(), snapshot.AddDate(0, 0, 7).UnixMilli()-1)
			res, err = http.Get(url)
			if err != nil {
				log.Fatal("http.Get error | ", err)
			}
			defer res.Body.Close()
			msg, err = io.ReadAll(res.Body)
			if err != nil {
				log.Fatal("io.ReadAll error | ", err)
			} // #endregion

			// #region Unmarshal api response to slice of interface slices
			var respInfc [][]interface{}
			err = json.Unmarshal(msg, &respInfc)
			if err != nil {
				log.Fatal("json.Unmarshal error | ", err)
			}
			if len(respInfc) < 21 {
				log.Println("Skipping entry. Not enough data for symbol at snapshot date | ", symbol, snapshot)
				continue
			} // #endregion

			// Iterate over api response slice and build slice to queue data for batch insert
			for _, markResp := range respInfc {
				log.Printf("symbol: %s snapshot_date: %s funding_time: %v mark_price: %s", symbol, snapshot, markResp[0].(int64), markResp[1].(string))
			}

			// [0] is funding_time, [1] is mark_price (open price)
			// add to batch insert slice where snapshot and symbol = snapshot and symbol

			// Sleep to prevent rate limiting
			time.Sleep(time.Millisecond * 25) // from binance api: weight = 1 for limit [1,100]. 2400 weight/min = 40 queries/sec
		}
	}

	// after "topN funding rates" table caught up to last entry of marketcap_snapshots
	// create new "funding averages" table foreign key fundingTime
	// calculate and insert average and median funding rates for fundingTime
	// calculate marketcap weighted ex-BTC average and median funding rates
	// calculate rolling 1 month 3 month 1 year funding rate confidence interval
	// insert all data to new "funding averages" table
	// after "funding averages" table caught up to last entry
	// query to find all fundingTime where average funding rates are at extremes
}
