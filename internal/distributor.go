package internal

import (
	"errors"
	"fmt"
	"sort"

	"github.com/gnolang/gno/gnoland"
	"github.com/gnolang/gno/pkgs/amino"
	"github.com/gnolang/gno/pkgs/bft/rpc/client"
	"github.com/gnolang/gno/pkgs/crypto/keys"
	"github.com/gnolang/gno/pkgs/sdk/bank"
	"github.com/gnolang/gno/pkgs/std"
	"go.uber.org/zap"
)

const (
	denomination = "ugnot"
)

var (
	// TODO support estimating gas params
	// These are constants for now,
	// but should be fetched as estimations
	// from the Tendermint node once this functionality
	// is available
	defaultGasFee = std.Coin{
		Denom:  denomination,
		Amount: 1,
	}

	// Each package call / deployment
	// costs a fixed 1 GNOT
	// https://github.com/gnolang/gno/issues/649
	initialTxCost = std.Coin{
		Denom:  denomination,
		Amount: 1000000, // 1 GNOT
	}
)

// distributor is the process
// that manages sub-account distributions
type distributor struct {
	logger *zap.Logger

	cli     client.Client
	keybase keys.Keybase
}

// newDistributor creates a new instance of the distributor
func newDistributor(
	logger *zap.Logger,
	cli client.Client,
	keybase keys.Keybase,
) *distributor {
	return &distributor{
		logger:  logger.Named("distributor"),
		cli:     cli,
		keybase: keybase,
	}
}

// distribute distributes the funds from the base account
// (account 0 in the mnemonic) to other subaccounts
func (d *distributor) distribute(
	subAccounts uint64,
	transactions uint64,
	mnemonic string,
) ([]keys.Info, error) {
	// Register the accounts with the keybase
	accounts := make([]keys.Info, subAccounts+1)

	for i := 0; i < int(subAccounts)+1; i++ {
		info, err := d.keybase.CreateAccount(
			"DistributorKey",
			mnemonic,
			"",
			encryptPassword,
			uint32(0),
			uint32(i),
		)
		if err != nil {
			return nil, err
		}

		accounts[i] = info
	}

	// Calculate the base fees
	subAccountCost := calculateRuntimeCosts(int64(transactions))

	return d.fundAccounts(accounts, subAccountCost)
}

// calculateRuntimeCosts calculates the amount of funds
// each account needs to have in order to participate in the
// stress test run
func calculateRuntimeCosts(totalTx int64) std.Coin {
	// Cost of a single run transaction for the sub-account
	// NOTE: Since there is no gas estimation support yet, this value
	// is fixed, but it will change in the future once pricing estimations
	// are added
	baseTxCost := defaultGasFee.Add(initialTxCost)

	// Each account should have enough funds
	// to execute the entire run
	subAccountCost := std.Coin{
		Denom:  denomination,
		Amount: totalTx * baseTxCost.Amount,
	}

	return subAccountCost
}

func (d *distributor) fundAccounts(accounts []keys.Info, singleRunCost std.Coin) ([]keys.Info, error) {
	type shortAccount struct {
		account      keys.Info
		missingFunds std.Coin
	}

	readyAccounts := make([]keys.Info, 0)

	// Find short accounts
	shortAccounts := make([]shortAccount, 0)

	for _, account := range accounts[1:] {
		// Fetch the account balance
		subAccount, err := d.getAccount(account.GetAddress().String())
		if err != nil {
			return nil, fmt.Errorf("unable to fetch sub-account, %w", err)
		}

		// Check if it has enough funds for the run
		if subAccount.Coins.AmountOf(denomination) < singleRunCost.Amount {
			shortAccounts = append(shortAccounts, shortAccount{
				account: account,
				missingFunds: std.Coin{
					Denom:  denomination,
					Amount: singleRunCost.Amount - subAccount.Coins.AmountOf(denomination),
				},
			})

			continue
		}

		readyAccounts = append(readyAccounts, account)
	}

	// Sort the short accounts so the ones with
	// the lowest missing funds are funded first
	sort.Slice(shortAccounts, func(i, j int) bool {
		return shortAccounts[i].missingFunds.IsLT(shortAccounts[j].missingFunds)
	})

	// Figure out how many accounts can be funded
	distributor, err := d.getAccount(accounts[0].GetAddress().String())
	if err != nil {
		return nil, fmt.Errorf("unable to fetch distributor account, %w", err)
	}

	distributorBalance := distributor.Coins
	fundableIndex := 0

	for _, account := range shortAccounts {
		// The transfer cost is the single run cost (missing balance) + 1ugnot fee (fixed)
		transferCost := std.NewCoins(defaultGasFee.Add(account.missingFunds))

		if distributorBalance.IsAllLT(transferCost) {
			// Distributor does not have any more funds
			// to cover the run cost
			break
		}

		fundableIndex++

		distributorBalance.Sub(transferCost)
	}

	if fundableIndex == 0 {
		// The distributor does not have funds to fund
		// any account for the stress test
		return nil, errors.New("insufficient distributor funds")
	}

	nonce := distributor.Sequence

	for _, account := range shortAccounts {
		msg := bank.MsgSend{
			FromAddress: distributor.GetAddress(),
			ToAddress:   account.account.GetAddress(),
			Amount:      std.NewCoins(account.missingFunds),
		}

		tx := std.Tx{
			Msgs: []std.Msg{
				msg,
			},
			Fee: std.NewFee(60000, defaultGasFee),
		}

		signers := tx.GetSigners()
		if tx.Signatures == nil {
			for range signers {
				tx.Signatures = append(tx.Signatures, std.Signature{
					PubKey:    nil, // zero signature
					Signature: nil, // zero signature
				})
			}
		}

		signature, pub, err := d.keybase.Sign(
			distributor.GetAddress().String(),
			encryptPassword,
			tx.GetSignBytes("dev", distributor.AccountNumber, nonce),
		)
		if err != nil {
			return nil, fmt.Errorf("unable to sign transaction, %w", err)
		}

		nonce++

		addr := pub.Address()
		found := false

		for i := range tx.Signatures {
			if signers[i] == addr {
				found = true
				tx.Signatures[i] = std.Signature{
					PubKey:    pub,
					Signature: signature,
				}
			}
		}

		if !found {
			return nil, fmt.Errorf("unable to sign transaction with address %s", distributor.GetAddress())
		}

		marshalledTx, err := amino.Marshal(tx)
		if err != nil {
			return nil, fmt.Errorf("unable to marshal transaction, %w", err)
		}

		res, err := d.cli.BroadcastTxCommit(marshalledTx)
		if err != nil {
			return nil, fmt.Errorf("unable to broadcast transfer transaction, %w", err)
		}

		if res.CheckTx.IsErr() {
			return nil, fmt.Errorf("transfer transaction check failed, %w", res.CheckTx.Error)
		}

		if res.DeliverTx.IsErr() {
			return nil, fmt.Errorf("transfer transaction delivery failed, %w", res.DeliverTx.Error)
		}

		readyAccounts = append(readyAccounts, account.account)
	}

	return readyAccounts, nil
}

// getAccount fetches account data from the node
func (d *distributor) getAccount(address string) (*gnoland.GnoAccount, error) {
	queryResult, err := d.cli.ABCIQuery(
		fmt.Sprintf("auth/accounts/%s", address),
		[]byte{},
	)

	if err != nil {
		return nil, fmt.Errorf("unable to fetch account %s, %w", address, err)
	}

	if queryResult.Response.IsErr() {
		return nil, fmt.Errorf("invalid account query result, %w", queryResult.Response.Error)
	}

	var acc gnoland.GnoAccount
	if err := amino.UnmarshalJSON(queryResult.Response.Data, &acc); err != nil {
		return nil, fmt.Errorf("unable to unmarshal query response, %w", err)
	}

	return &acc, nil
}
