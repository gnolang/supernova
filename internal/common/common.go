package common

import "github.com/gnolang/gno/tm2/pkg/std"

const Denomination = "ugnot"

var DefaultGasFee = std.Coin{
	Denom:  Denomination,
	Amount: 5,
}
