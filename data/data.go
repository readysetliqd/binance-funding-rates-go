package data

import (
	"database/sql"
	"time"
)

type Symbol struct {
	Symbol string
	Rank   int64
}

type Row struct {
	FundingTime  int64
	Symbol       string
	FundingRate  float64
	MarkPrice    sql.NullFloat64
	SnapshotDate time.Time
	Rank         int64
}

type FundingRateApiResp struct {
	Symbol string `json:"symbol"`
	Time   int64  `json:"fundingTime"`
	Rate   string `json:"fundingRate"`
	Mark   string `json:"markPrice"`
}

var StableCoins = []string{
	"'BUSD'",
	"'BITEUR'",
	"'BITUSD'",
	"'DAI'",
	"'EURS'",
	"'HUSD'",
	"'LUSD'",
	"'PAX'",
	"'RAI'",
	"'TUSD'",
	"'USDC'",
	"'USDD'",
	"'USDN'",
	"'USDP'",
	"'USDT'",
	"'UST'",
	"'VAI'",
	"'XUSD'",
}

var SymbolsBefore2020 = []string{"BTC", "ETH", "BCH"}

var SymbolsBefore2021 = []string{
	"BTC", "ETH", "BCH", "XRP", "LTC", "BNB", "LINK", "ADA", "DOT", "XLM",
	"XMR", "EOS", "TRX", "XTZ", "THETA", "NEO", "DASH", "VET", "ATOM", "FIL",
	"UNI", "AAVE", "SNX", "ZIL", "IOTA", "ZEC", "YFI", "WAVES", "ETC", "DOGE",
	"COMP", "MKR", "GRT", "SUSHI", "KSM", "ALGO", "OMG", "ONT", "EGLD", "BAT",
	"ZRX", "REN", "NEAR", "ICX", "QTUM", "AVAX", "REP", "RSR", "KNC", "RUNE",
	"OCEAN", "ZEN", "CHZ", "BNT", "ENJ", "BAND", "BAL", "IOST", "MATIC", "HNT",
	"CRV", "STORJ", "RLC", "SOL", "KAVA", "CVC", "SXP", "SRM", "YFII", "TOMO",
	"FTM", "ALPHA", "AXS", "LEND", "UNFI", "SKL", "BEL", "CTK", "FLM", "BLZ",
	"BZRX",
}

var SymbolsBefore2022 = []string{
	"BTC", "ETH", "BCH", "XRP", "LTC", "BNB", "LINK", "ADA", "DOT", "XLM",
	"XMR", "EOS", "TRX", "XTZ", "THETA", "NEO", "DASH", "VET", "ATOM", "FIL",
	"UNI", "AAVE", "SNX", "ZIL", "IOTA", "ZEC", "YFI", "WAVES", "ETC", "DOGE",
	"COMP", "MKR", "GRT", "SUSHI", "KSM", "ALGO", "OMG", "ONT", "EGLD", "BAT",
	"ZRX", "REN", "NEAR", "ICX", "QTUM", "AVAX", "REP", "RSR", "KNC", "RUNE",
	"OCEAN", "ZEN", "CHZ", "BNT", "ENJ", "BAND", "BAL", "IOST", "MATIC", "HNT",
	"CRV", "STORJ", "RLC", "SOL", "KAVA", "CVC", "SXP", "SRM", "YFII", "TOMO",
	"FTM", "ALPHA", "AXS", "LEND", "UNFI", "SKL", "BEL", "CTK", "FLM", "BLZ",
	"BZRX", "AKRO", "SAND", "ANKR", "REEF", "RVN", "SFP", "DODO", "LIT", "LUNA",
	"COTI", "XEM", "1INCH", "CELR", "HOT", "DENT", "STMX", "LINA", "ONE", "ALICE",
	"CHR", "MANA", "HBAR", "DGB", "NKN", "SC", "BTT", "OGN", "MTL", "BAKE",
	"ICP", "1000SHIB", "GALA", "KLAY", "CELO", "AR", "CTSI", "ARPA", "NU", "LPT",
	"ENS", "PEOPLE", "ANT", "ROSE",
}

var SymbolsBefore2023 = []string{
	"BTC", "ETH", "BCH", "XRP", "LTC", "BNB", "LINK", "ADA", "DOT", "XLM",
	"XMR", "EOS", "TRX", "XTZ", "THETA", "NEO", "DASH", "VET", "ATOM", "FIL",
	"UNI", "AAVE", "SNX", "ZIL", "IOTA", "ZEC", "YFI", "WAVES", "ETC", "DOGE",
	"COMP", "MKR", "GRT", "SUSHI", "KSM", "ALGO", "OMG", "ONT", "EGLD", "BAT",
	"ZRX", "REN", "NEAR", "ICX", "QTUM", "AVAX", "REP", "RSR", "KNC", "RUNE",
	"OCEAN", "ZEN", "CHZ", "BNT", "ENJ", "BAND", "BAL", "IOST", "MATIC", "HNT",
	"CRV", "STORJ", "RLC", "SOL", "KAVA", "CVC", "SXP", "SRM", "YFII", "TOMO",
	"FTM", "ALPHA", "AXS", "LEND", "UNFI", "SKL", "BEL", "CTK", "FLM", "BLZ",
	"BZRX", "AKRO", "SAND", "ANKR", "REEF", "RVN", "SFP", "DODO", "LIT", "LUNA",
	"COTI", "XEM", "1INCH", "CELR", "HOT", "DENT", "STMX", "LINA", "ONE", "ALICE",
	"CHR", "MANA", "HBAR", "DGB", "NKN", "SC", "BTT", "OGN", "MTL", "BAKE",
	"ICP", "1000SHIB", "GALA", "KLAY", "CELO", "AR", "CTSI", "ARPA", "NU", "LPT",
	"ENS", "PEOPLE", "ANT", "ROSE", "DUSK", "ANC", "API3", "IMX", "FLOW", "WOO",
	"BNX", "APE", "GMT", "GAL", "DAR", "JASMY", "OP", "INJ", "CVX", "LDO",
	"PHB", "AMB", "LUNA2", "1000LUNC", "SPELL", "STG", "APT", "QNT", "FTT",
}
