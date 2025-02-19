package common

import "github.com/gnolang/gno/tm2/pkg/std"

const Denomination = "ugnot"

// DefaultGasPrice represents the gno.land chain's
// default minimum gas price ratio, which is 0.001ugnot/gas
var DefaultGasPrice = std.GasPrice{
	Gas: 1000,
	Price: std.Coin{
		Denom:  Denomination,
		Amount: 1,
	},
}

// CalculateFeeInRatio calculates the minimum gas fee that should be specified
// in a transaction, given the gas wanted (of the tx) and the reference gas ratio
func CalculateFeeInRatio(gasWanted int64, reference std.GasPrice) std.Fee {
	// required amount = ceil((gas wanted * reference.Price.Amount) / reference.Gas)
	requiredAmount := (gasWanted*reference.Price.Amount + reference.Gas - 1) / reference.Gas

	return std.Fee{
		GasWanted: gasWanted,
		GasFee: std.Coin{
			Denom:  reference.Price.Denom,
			Amount: requiredAmount,
		},
	}
}
