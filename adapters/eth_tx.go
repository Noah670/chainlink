package adapters

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/store"
	"github.com/smartcontractkit/chainlink/store/models"
	"github.com/smartcontractkit/chainlink/utils"
)

type EthTx struct {
	Address    common.Address    `json:"address"`
	FunctionID models.FunctionID `json:"functionID"`
}

func (etx *EthTx) Perform(input models.RunResult, store *store.Store) models.RunResult {
	if !input.Pending {
		return createTxRunResult(etx, input, store)
	} else {
		return ensureTxRunResult(input, store)
	}
}

func createTxRunResult(
	e *EthTx,
	input models.RunResult,
	store *store.Store,
) models.RunResult {
	data, err := hex.DecodeString(e.FunctionID.WithoutPrefix() + input.Value())
	if err != nil {
		return models.RunResultWithError(err)
	}

	attempt, err := store.TxManager.CreateTx(e.Address, data)
	if err != nil {
		return models.RunResultWithError(err)
	}

	sendResult := models.RunResultWithValue(attempt.Hash.String())
	return ensureTxRunResult(sendResult, store)
}

func ensureTxRunResult(input models.RunResult, store *store.Store) models.RunResult {
	hash, err := utils.StringToHash(input.Value())
	if err != nil {
		return models.RunResultWithError(err)
	}

	confirmed, err := store.TxManager.EnsureTxConfirmed(hash)

	if err != nil {
		return models.RunResultWithError(err)
	} else if !confirmed {
		return models.RunResultPending(input)
	}
	return models.RunResultWithValue(hash.String())
}