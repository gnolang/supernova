package common

import "github.com/gnolang/gno/tm2/pkg/std"

const Denomination = "ugnot"

var (
	DefaultGasFee = std.Coin{
		Denom:  Denomination,
		Amount: 5,
	}

	// TODO remove
	InitialTxCost = std.Coin{
		Denom:  Denomination,
		Amount: 1000000, // 1 GNOT
	}
)
