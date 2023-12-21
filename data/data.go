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
	"'USTC'",
	"'VAI'",
	"'XUSD'",
}

var ThousandSymbols = []string{
	"BONK",
	"FLOKI",
	"LUNC",
	"PEPE",
	"RATS",
	"SATS",
	"SHIB",
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

var SymbolsBefore2024 = []string{
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
	"PHB", "AMB", "LUNA2", "1000LUNC", "SPELL", "STG", "APT", "QNT", "FTT", "FET",
	"FXS", "HOOK", "MAGIC", "T", "RNDR", "MINA", "HIGH", "ASTR", "AGIX", "GMX",
	"CFX", "STX", "COCOS", "ACH", "SSV", "CKB", "PERP", "TRU", "LQTY", "ARB",
	"ID", "JOE", "LEVER", "TLM", "RDNT", "HFT", "XVS", "BLUR", "EDU", "IDEX",
	"SUI", "1000PEPE", "1000FLOKI", "RAD", "UMA", "KEY", "COMBO", "NMR", "MAV", "MDT",
	"XVG", "WLD", "PENDLE", "ARKM", "AGLD", "YGG", "OXT", "SEI", "CYBER", "HIFI",
	"ARK", "FRONT", "GLMR", "BICO", "STRAX", "LOOM", "BIGTIME", "BOND", "ORBS", "STPT",
	"WAXP", "BSV", "RIF", "POLYX", "GAS", "POWR", "SLP", "TIA", "SNT", "CAKE",
	"TWT", "MEME", "TOKEN", "ORDI", "STEEM", "BADGER", "ILV", "NTRN", "MBL", "KAS",
	"BEAMX", "1000BONK", "PYTH", "SUPER", "USTC", "USDC", "ONG", "ETHW", "JTO", "1000SATS",
	"AUCTION", "1000RATS", "ACE",
}
