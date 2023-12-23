# binance-funding-rates-go
## Description
Program that builds a table in database for historical funding rates of the topN number of coins by market cap from binance public api and analyzes the aggregated funding rates. Intent is to test predictiability of forward returns of equal weighted longs of all topN coins at any given point of "extreme" aggregated funding rates. 

Note: Must be used with database and table built from github.com/readysetliqd/crypto-historical-marketcaps-scraper-go

## Requirements
- Go 1.21.3
- PostgreSQL 14
    - Existing database and table built from crypto-historical-marketcaps-scraper-go
- Python 3.9.13

## Python Libraries
- Install with cmd ```pip install -r requirements.txt``` as required

## Setup
- Install any necessary requirements
- Rename db_sample.env to db.env
- Fill in system specific sensitive data database access
    - (optional) Copy filled db.env file and paste into this directory from your clone of crypto-historical-marketcaps-scraper-go
- (optional) Edit configs at the top of main.go file as desired
    - Change topN to desired number of coins to pull data for. Keep in mind this will get the top number of existing coins on binance futures in order by market cap. Since not all the coins in the top eg. 100 on CoinMarketCap have always been listed on Binance Futures, the program will keep pulling data for coins until topN number is reached
    - Ensure snapshotsTableName matches table name already existing in your database from crypto-historical-marketcaps-scraper-go
    - Recommended to leave some call to strconv.Itoa(topN) in fundingTableName in case you run this program with multiple different topN values
- Run main.go to build table in database and fill data
- Run python-averages-rolling-windows.py
- See newly created stats_output.txt for results