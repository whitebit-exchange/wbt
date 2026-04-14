package core

import (
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func TestWbtMainnetTransactionToMessageRejectsUnprotectedLegacyTx(t *testing.T) {
	t.Parallel()
	key, err := crypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	tx, err := types.SignTx(
		types.NewTransaction(0, crypto.PubkeyToAddress(key.PublicKey), big.NewInt(1), params.TxGas, big.NewInt(1), nil),
		types.HomesteadSigner{},
		key,
	)
	if err != nil {
		t.Fatalf("sign tx: %v", err)
	}
	if tx.Protected() {
		t.Fatal("expected unprotected legacy tx")
	}
	signer := types.MakeSigner(params.WbtMainnetChainConfig, big.NewInt(1))
	_, err = TransactionToMessage(tx, signer, nil)
	if !errors.Is(err, types.ErrTxTypeNotSupported) {
		t.Fatalf("expected ErrTxTypeNotSupported, got %v", err)
	}
}
