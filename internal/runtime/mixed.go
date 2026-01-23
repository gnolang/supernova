package runtime

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/gnolang/gno/gno.land/pkg/sdk/vm"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/supernova/internal/common"
	"github.com/gnolang/supernova/internal/signer"
	"github.com/schollz/progressbar/v3"
)

type mixedRuntime struct {
	ctx       context.Context
	config    *MixConfig
	realmPath string
}

func newMixedRuntime(ctx context.Context, config *MixConfig) *mixedRuntime {
	return &mixedRuntime{
		ctx:    ctx,
		config: config,
	}
}

func (m *mixedRuntime) Initialize(
	account std.Account,
	signFn SignFn,
	estimateFn EstimateGasFn,
	currentMaxGas int64,
	gasPrice std.GasPrice,
) ([]*std.Tx, error) {
	if !m.config.HasType(RealmCall) {
		return nil, nil
	}

	m.realmPath = fmt.Sprintf(
		"%s/%s/stress_%d",
		realmPathPrefix,
		account.GetAddress().String(),
		time.Now().Unix(),
	)

	msg := vm.MsgAddPackage{
		Creator: account.GetAddress(),
		Package: &std.MemPackage{
			Name: packageName,
			Path: m.realmPath,
			Files: []*std.MemFile{
				{
					Name: gnomodFileName,
					Body: gnomodBody,
				},
				{
					Name: realmFileName,
					Body: realmBody,
				},
			},
		},
	}

	tx := &std.Tx{
		Msgs: []std.Msg{msg},
		Fee:  common.CalculateFeeInRatio(currentMaxGas, gasPrice),
	}

	err := signFn(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to sign initialize transaction, %w", err)
	}

	gasWanted, err := estimateFn(m.ctx, tx)
	if err != nil {
		return nil, fmt.Errorf("unable to estimate gas: %w", err)
	}

	tx.Signatures = make([]std.Signature, 0)
	tx.Fee = common.CalculateFeeInRatio(gasWanted+gasBuffer, gasPrice)

	err = signFn(tx)
	if err != nil {
		return nil, fmt.Errorf("unable to sign initialize transaction, %w", err)
	}

	return []*std.Tx{tx}, nil
}

func (m *mixedRuntime) CalculateRuntimeCosts(
	account std.Account,
	estimateFn EstimateGasFn,
	signFn SignFn,
	currentMaxGas int64,
	gasPrice std.GasPrice,
	transactions uint64,
) (std.Coin, error) {
	fmt.Printf("\n⏳ Estimating Gas ⏳\n")

	txCounts := m.config.CalculateTransactionCounts(transactions)
	var totalGas int64

	for txType, count := range txCounts {
		if count == 0 {
			continue
		}

		txFee := common.CalculateFeeInRatio(currentMaxGas, gasPrice)
		msg := m.getMsgForType(txType, account, 0)

		tx := &std.Tx{
			Msgs: []std.Msg{msg},
			Fee:  txFee,
		}

		err := signFn(tx)
		if err != nil {
			return std.Coin{}, fmt.Errorf("unable to sign transaction for %s, %w", txType, err)
		}

		estimatedGas, err := estimateFn(m.ctx, tx)
		if err != nil {
			return std.Coin{}, fmt.Errorf("unable to estimate gas for %s, %w", txType, err)
		}

		totalGas += int64(count) * estimatedGas
	}

	return std.Coin{
		Denom:  common.Denomination,
		Amount: totalGas,
	}, nil
}

func (m *mixedRuntime) ConstructTransactions(
	keys []crypto.PrivKey,
	accounts []std.Account,
	transactions uint64,
	maxGas int64,
	gasPrice std.GasPrice,
	chainID string,
	estimateFn EstimateGasFn,
) ([]*std.Tx, error) {
	txCounts := m.config.CalculateTransactionCounts(transactions)
	typeSequence := m.generateShuffledSequence(txCounts)

	gasEstimates := make(map[Type]int64)

	fmt.Printf("\n⏳ Estimating Gas Per Type ⏳\n")

	for txType, count := range txCounts {
		if count == 0 {
			continue
		}

		txFee := common.CalculateFeeInRatio(maxGas, gasPrice)
		msg := m.getMsgForType(txType, accounts[0], 0)

		tx := &std.Tx{
			Msgs: []std.Msg{msg},
			Fee:  txFee,
		}

		cfg := signer.SignCfg{
			ChainID:       chainID,
			AccountNumber: accounts[0].GetAccountNumber(),
			Sequence:      accounts[0].GetSequence(),
		}

		if err := signer.SignTx(tx, keys[0], cfg); err != nil {
			return nil, fmt.Errorf("unable to sign transaction for %s, %w", txType, err)
		}

		gasWanted, err := estimateFn(m.ctx, tx)
		if err != nil {
			return nil, fmt.Errorf("unable to estimate gas for %s, %w", txType, err)
		}

		gasEstimates[txType] = gasWanted + gasBuffer
		fmt.Printf("Estimated Gas for %s: %d\n", txType, gasEstimates[txType])
	}

	fmt.Printf("\n🔨 Constructing Transactions 🔨\n\n")

	txs := make([]*std.Tx, transactions)
	nonceMap := make(map[uint64]uint64)
	typeCounters := make(map[Type]int)

	bar := progressbar.Default(int64(transactions), "constructing txs")

	for i, txType := range typeSequence {
		creator := accounts[i%len(accounts)]
		creatorKey := keys[i%len(accounts)]
		accountNumber := creator.GetAccountNumber()

		typeIndex := typeCounters[txType]
		typeCounters[txType]++

		msg := m.getMsgForType(txType, creator, typeIndex)
		txFee := common.CalculateFeeInRatio(gasEstimates[txType], gasPrice)

		tx := &std.Tx{
			Msgs: []std.Msg{msg},
			Fee:  txFee,
		}

		nonce, found := nonceMap[accountNumber]
		if !found {
			nonce = creator.GetSequence()
			nonceMap[accountNumber] = nonce
		}

		cfg := signer.SignCfg{
			ChainID:       chainID,
			AccountNumber: accountNumber,
			Sequence:      nonce,
		}

		if err := signer.SignTx(tx, creatorKey, cfg); err != nil {
			return nil, fmt.Errorf("unable to sign transaction, %w", err)
		}

		nonceMap[accountNumber] = nonce + 1
		txs[i] = tx
		_ = bar.Add(1)
	}

	fmt.Printf("✅ Successfully constructed %d transactions\n", transactions)

	return txs, nil
}

func (m *mixedRuntime) generateShuffledSequence(txCounts map[Type]uint64) []Type {
	var sequence []Type
	for txType, count := range txCounts {
		for i := uint64(0); i < count; i++ {
			sequence = append(sequence, txType)
		}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	rng.Shuffle(len(sequence), func(i, j int) {
		sequence[i], sequence[j] = sequence[j], sequence[i]
	})

	return sequence
}

func (m *mixedRuntime) getMsgForType(txType Type, creator std.Account, index int) std.Msg {
	timestamp := time.Now().Unix()

	switch txType {
	case RealmCall:
		return vm.MsgCall{
			Caller:  creator.GetAddress(),
			PkgPath: m.realmPath,
			Func:    methodName,
			Args:    []string{fmt.Sprintf("Account-%d", index)},
		}
	case RealmDeployment:
		return vm.MsgAddPackage{
			Creator: creator.GetAddress(),
			Package: &std.MemPackage{
				Name: packageName,
				Path: fmt.Sprintf(
					"%s/%s/stress_%d_%d",
					realmPathPrefix,
					creator.GetAddress().String(),
					timestamp,
					index,
				),
				Files: []*std.MemFile{
					{
						Name: gnomodFileName,
						Body: gnomodBody,
					},
					{
						Name: realmFileName,
						Body: realmBody,
					},
				},
			},
		}
	case PackageDeployment:
		return vm.MsgAddPackage{
			Creator: creator.GetAddress(),
			Package: &std.MemPackage{
				Name: packageName,
				Path: fmt.Sprintf(
					"%s/%s/stress_%d_%d",
					packagePathPrefix,
					creator.GetAddress().String(),
					timestamp,
					index,
				),
				Files: []*std.MemFile{
					{
						Name: gnomodFileName,
						Body: gnomodBody,
					},
					{
						Name: packageFileName,
						Body: packageBody,
					},
				},
			},
		}
	default:
		return nil
	}
}
