package clique

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus/misc"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

func TestCliqueRejectsMalleatedSeal(t *testing.T) {
	var (
		db      = rawdb.NewMemoryDatabase()
		key, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		signer  = crypto.PubkeyToAddress(key.PublicKey)
		config  = *params.AllCliqueProtocolChanges
		cliqueC = *params.AllCliqueProtocolChanges.Clique
	)

	cliqueC.Period = 1
	config.Clique = &cliqueC
	config.CepheusBlock = big.NewInt(0)
	genesis := &core.Genesis{
		Config:    &config,
		ExtraData: make([]byte, extraVanity+common.AddressLength+extraSeal),
		GasLimit:  10_000_000,
		BaseFee:   big.NewInt(params.InitialBaseFee),
		Alloc: map[common.Address]core.GenesisAccount{
			signer: {Balance: big.NewInt(1_000_000_000_000_000_000)},
		},
	}

	copy(genesis.ExtraData[extraVanity:], signer[:])
	engine := New(config.Clique, db)
	chain, err := core.NewBlockChain(db, nil, genesis, nil, engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("failed to create clique chain: %v", err)
	}

	defer chain.Stop()
	parent := chain.CurrentHeader()
	original := &types.Header{
		ParentHash:  parent.Hash(),
		UncleHash:   types.EmptyUncleHash,
		Coinbase:    common.Address{},
		Root:        parent.Root,
		TxHash:      types.EmptyTxsHash,
		ReceiptHash: types.EmptyReceiptsHash,
		Bloom:       types.Bloom{},
		Difficulty:  new(big.Int).Set(diffInTurn),
		Number:      new(big.Int).Add(parent.Number, common.Big1),
		GasLimit:    parent.GasLimit,
		GasUsed:     0,
		Time:        parent.Time + config.Clique.Period,
		Extra:       make([]byte, extraVanity+extraSeal),
		MixDigest:   common.Hash{},
		Nonce:       types.BlockNonce{},
		BaseFee:     misc.CalcBaseFee(&config, parent),
	}

	sig, err := crypto.Sign(SealHash(original).Bytes(), key)
	if err != nil {
		t.Fatalf("failed to sign clique header: %v", err)
	}

	copy(original.Extra[len(original.Extra)-extraSeal:], sig)
	forged := types.CopyHeader(original)
	malleated := malleateSignature(sig)
	copy(forged.Extra[len(forged.Extra)-extraSeal:], malleated)

	if original.Hash() == forged.Hash() {
		t.Fatal("expected different block hashes for original and forged seal")
	}

	if SealHash(original) != SealHash(forged) {
		t.Fatalf("expected identical seal hash, have %s want %s", SealHash(forged), SealHash(original))
	}

	if err := engine.VerifyHeader(chain, original, true); err != nil {
		t.Fatalf("original header rejected: %v", err)
	}

	if err := engine.VerifyHeader(chain, forged, true); err == nil {
		t.Fatal("expected malleated header to be rejected")
	}

	if SealHash(original) != SealHash(forged) {
		t.Fatalf("expected identical seal hash, have %s want %s", SealHash(forged), SealHash(original))
	}

	authorOriginal, err := engine.Author(original)
	if err != nil {
		t.Fatalf("failed to recover original signer: %v", err)
	}

	if authorOriginal != signer {
		t.Fatalf("expected original header to recover signer %s, got %s", signer, authorOriginal)
	}
}

func TestCliqueRejectsPreShanghaiHeaderWithWithdrawalsRoot(t *testing.T) {
	t.Parallel()
	var (
		db      = rawdb.NewMemoryDatabase()
		key, _  = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		signer  = crypto.PubkeyToAddress(key.PublicKey)
		config  = *params.AllCliqueProtocolChanges
		cliqueC = *params.AllCliqueProtocolChanges.Clique
	)

	cliqueC.Period = 1
	config.Clique = &cliqueC
	config.CepheusBlock = big.NewInt(0)
	genesis := &core.Genesis{
		Config:    &config,
		ExtraData: make([]byte, extraVanity+common.AddressLength+extraSeal),
		GasLimit:  10_000_000,
		BaseFee:   big.NewInt(params.InitialBaseFee),
		Alloc: map[common.Address]core.GenesisAccount{
			signer: {Balance: big.NewInt(1_000_000_000_000_000_000)},
		},
	}

	copy(genesis.ExtraData[extraVanity:], signer[:])
	engine := New(config.Clique, db)
	chain, err := core.NewBlockChain(db, nil, genesis, nil, engine, vm.Config{}, nil, nil)
	if err != nil {
		t.Fatalf("failed to create clique chain: %v", err)
	}

	defer chain.Stop()
	parent := chain.CurrentHeader()
	if config.IsShanghai(parent.Time) {
		t.Fatal("test expects pre-Shanghai chain config (AllCliqueProtocolChanges)")
	}

	wh := types.EmptyWithdrawalsHash
	crafted := &types.Header{
		ParentHash:      parent.Hash(),
		UncleHash:       types.EmptyUncleHash,
		Coinbase:        common.Address{},
		Root:            parent.Root,
		TxHash:          types.EmptyTxsHash,
		ReceiptHash:     types.EmptyReceiptsHash,
		Bloom:           types.Bloom{},
		Difficulty:      new(big.Int).Set(diffInTurn),
		Number:          new(big.Int).Add(parent.Number, common.Big1),
		GasLimit:        parent.GasLimit,
		GasUsed:         0,
		Time:            parent.Time + config.Clique.Period,
		Extra:           make([]byte, extraVanity+extraSeal),
		MixDigest:       common.Hash{},
		Nonce:           types.BlockNonce{},
		BaseFee:         misc.CalcBaseFee(&config, parent),
		WithdrawalsHash: &wh,
	}

	// Any 65-byte ECDSA signature with valid (v,r,s) reaches ecrecover → SealHash
	// on vulnerable code; we only need to prove VerifyHeader does not panic.
	sig, err := crypto.Sign(make([]byte, 32), key)
	if err != nil {
		t.Fatalf("failed to create signature bytes: %v", err)
	}
	copy(crafted.Extra[len(crafted.Extra)-extraSeal:], sig)

	err = engine.VerifyHeader(chain, crafted, true)
	if err == nil || err.Error() != "unexpected withdrawals root in clique header" {
		t.Fatalf("expected withdrawals-root rejection error, got %v", err)
	}
}

func malleateSignature(sig []byte) []byte {
	out := append([]byte(nil), sig...)
	s := new(big.Int).SetBytes(out[32:64])
	s.Sub(crypto.S256().Params().N, s)
	sBytes := s.FillBytes(make([]byte, 32))
	copy(out[32:64], sBytes)
	out[64] ^= 1

	return out
}
