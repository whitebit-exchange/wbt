package txpool

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

// TestWbtMainnetTxPoolRejectsUnprotectedLegacyTransactions ensures unprotected
// legacy txs are rejected on remote ingestion for the WhiteBIT mainnet chain.
func TestWbtMainnetTxPoolRejectsUnprotectedLegacyTransactions(t *testing.T) {
	t.Parallel()
	pool, key := setupPoolWithConfig(params.WbtMainnetChainConfig)
	defer pool.Stop()
	from := crypto.PubkeyToAddress(key.PublicKey)
	testAddBalance(pool, from, new(big.Int).Mul(big.NewInt(params.Ether), big.NewInt(10)))
	tx, err := types.SignTx(
		types.NewTransaction(0, from, big.NewInt(1), params.TxGas, big.NewInt(1), nil),
		types.HomesteadSigner{},
		key,
	)

	if err != nil {
		t.Fatalf("failed to sign unprotected tx: %v", err)
	}

	if tx.Protected() {
		t.Fatal("expected test transaction to be unprotected")
	}

	errs := pool.AddRemotesSync([]*types.Transaction{tx})
	if len(errs) != 1 {
		t.Fatalf("unexpected error slice length: %d", len(errs))
	}

	if !errors.Is(errs[0], core.ErrTxTypeNotSupported) {
		t.Fatalf("expected ErrTxTypeNotSupported, got %v", errs[0])
	}

	pending, _ := pool.ContentFrom(from)
	if len(pending) != 0 {
		t.Fatalf("expected empty pending pool, got %d txs", len(pending))
	}
}
