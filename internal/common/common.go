package common

import "github.com/gnolang/gno/pkgs/std"

const (
	Denomination    = "ugnot"
	EncryptPassword = "encrypt"
	KeybasePrefix   = "stress-account-"
)

// TODO support estimating gas params
// These are constants for now,
// but should be fetched as estimations
// from the Tendermint node once this functionality
// is available.
//
// Each package call / deployment
// costs a fixed 1 GNOT
// https://github.com/gnolang/gno/issues/649
var (
	DefaultGasFee = std.Coin{
		Denom:  Denomination,
		Amount: 1,
	}

	InitialTxCost = std.Coin{
		Denom:  Denomination,
		Amount: 1000000, // 1 GNOT
	}
)
