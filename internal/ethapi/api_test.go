// Copyright 2023 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package ethapi

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ethereum/go-ethereum/trie"
	"github.com/stretchr/testify/require"
)

func TestTransaction_RoundTripRpcJSON(t *testing.T) {
	var (
		config = params.AllEthashProtocolChanges
		signer = types.LatestSigner(config)
		key, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
		tests  = allTransactionTypes(common.Address{0xde, 0xad}, config)
	)
	t.Parallel()
	for i, tt := range tests {
		var tx2 types.Transaction
		tx, err := types.SignNewTx(key, signer, tt)
		if err != nil {
			t.Fatalf("test %d: signing failed: %v", i, err)
		}
		// Regular transaction
		if data, err := json.Marshal(tx); err != nil {
			t.Fatalf("test %d: marshalling failed; %v", i, err)
		} else if err = tx2.UnmarshalJSON(data); err != nil {
			t.Fatalf("test %d: sunmarshal failed: %v", i, err)
		} else if want, have := tx.Hash(), tx2.Hash(); want != have {
			t.Fatalf("test %d: stx changed, want %x have %x", i, want, have)
		}

		//  rpcTransaction
		rpcTx := newRPCTransaction(tx, common.Hash{}, 0, 0, nil, config)
		if data, err := json.Marshal(rpcTx); err != nil {
			t.Fatalf("test %d: marshalling failed; %v", i, err)
		} else if err = tx2.UnmarshalJSON(data); err != nil {
			t.Fatalf("test %d: unmarshal failed: %v", i, err)
		} else if want, have := tx.Hash(), tx2.Hash(); want != have {
			t.Fatalf("test %d: tx changed, want %x have %x", i, want, have)
		}
	}
}

func allTransactionTypes(addr common.Address, config *params.ChainConfig) []types.TxData {
	return []types.TxData{
		&types.LegacyTx{
			Nonce:    5,
			GasPrice: big.NewInt(6),
			Gas:      7,
			To:       &addr,
			Value:    big.NewInt(8),
			Data:     []byte{0, 1, 2, 3, 4},
			V:        big.NewInt(9),
			R:        big.NewInt(10),
			S:        big.NewInt(11),
		},
		&types.LegacyTx{
			Nonce:    5,
			GasPrice: big.NewInt(6),
			Gas:      7,
			To:       nil,
			Value:    big.NewInt(8),
			Data:     []byte{0, 1, 2, 3, 4},
			V:        big.NewInt(32),
			R:        big.NewInt(10),
			S:        big.NewInt(11),
		},
		&types.AccessListTx{
			ChainID:  config.ChainID,
			Nonce:    5,
			GasPrice: big.NewInt(6),
			Gas:      7,
			To:       &addr,
			Value:    big.NewInt(8),
			Data:     []byte{0, 1, 2, 3, 4},
			AccessList: types.AccessList{
				types.AccessTuple{
					Address:     common.Address{0x2},
					StorageKeys: []common.Hash{types.EmptyRootHash},
				},
			},
			V: big.NewInt(32),
			R: big.NewInt(10),
			S: big.NewInt(11),
		},
		&types.AccessListTx{
			ChainID:  config.ChainID,
			Nonce:    5,
			GasPrice: big.NewInt(6),
			Gas:      7,
			To:       nil,
			Value:    big.NewInt(8),
			Data:     []byte{0, 1, 2, 3, 4},
			AccessList: types.AccessList{
				types.AccessTuple{
					Address:     common.Address{0x2},
					StorageKeys: []common.Hash{types.EmptyRootHash},
				},
			},
			V: big.NewInt(32),
			R: big.NewInt(10),
			S: big.NewInt(11),
		},
		&types.DynamicFeeTx{
			ChainID:   config.ChainID,
			Nonce:     5,
			GasTipCap: big.NewInt(6),
			GasFeeCap: big.NewInt(9),
			Gas:       7,
			To:        &addr,
			Value:     big.NewInt(8),
			Data:      []byte{0, 1, 2, 3, 4},
			AccessList: types.AccessList{
				types.AccessTuple{
					Address:     common.Address{0x2},
					StorageKeys: []common.Hash{types.EmptyRootHash},
				},
			},
			V: big.NewInt(32),
			R: big.NewInt(10),
			S: big.NewInt(11),
		},
		&types.DynamicFeeTx{
			ChainID:    config.ChainID,
			Nonce:      5,
			GasTipCap:  big.NewInt(6),
			GasFeeCap:  big.NewInt(9),
			Gas:        7,
			To:         nil,
			Value:      big.NewInt(8),
			Data:       []byte{0, 1, 2, 3, 4},
			AccessList: types.AccessList{},
			V:          big.NewInt(32),
			R:          big.NewInt(10),
			S:          big.NewInt(11),
		},
	}
}

type backend struct {
	*backendMock
	chainConfig    *params.ChainConfig
	getTransaction func(context.Context, common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error)
	blockByHash    func(context.Context, common.Hash) (*types.Block, error)
	getReceipts    func(context.Context, common.Hash) (types.Receipts, error)
}

func (b *backend) GetTransaction(ctx context.Context, hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
	return b.getTransaction(ctx, hash)
}

func (b *backend) BlockByHash(ctx context.Context, hash common.Hash) (*types.Block, error) {
	return b.blockByHash(ctx, hash)
}

func (b *backend) GetReceipts(ctx context.Context, hash common.Hash) (types.Receipts, error) {
	return b.getReceipts(ctx, hash)
}

func (b *backend) ChainConfig() *params.ChainConfig {
	return b.chainConfig
}

func TestGetBlockReceiptsFailures(t *testing.T) {
	testCases := map[string]struct {
		backend        *backend
		notEmptyResult bool
		expectedError  string
	}{
		"error when getting block by hash": {
			backend: &backend{
				backendMock: newBackendMock(),
				chainConfig: &params.ChainConfig{},
				blockByHash: func(_ context.Context, hash common.Hash) (*types.Block, error) {
					return nil, errors.New("failed to get block by hash")
				},
			},
			expectedError: "failed to get block by hash",
		},
		"block not found by hash": {
			backend: &backend{
				backendMock: newBackendMock(),
				chainConfig: &params.ChainConfig{},
				blockByHash: func(_ context.Context, hash common.Hash) (*types.Block, error) {
					return nil, nil
				},
			},
		},
		"error when getting receipts by block hash": {
			backend: &backend{
				backendMock: newBackendMock(),
				chainConfig: &params.ChainConfig{},
				blockByHash: func(_ context.Context, hash common.Hash) (*types.Block, error) {
					return &types.Block{}, nil
				},
				getReceipts: func(ctx context.Context, hash common.Hash) (types.Receipts, error) {
					return nil, errors.New("failed to get receipts by block hash")
				},
			},
			expectedError: "failed to get receipts by block hash",
		},
		"empty receipts list": {
			backend: &backend{
				backendMock: newBackendMock(),
				chainConfig: &params.ChainConfig{},
				blockByHash: func(_ context.Context, hash common.Hash) (*types.Block, error) {
					return &types.Block{}, nil
				},
				getReceipts: func(ctx context.Context, hash common.Hash) (types.Receipts, error) {
					return nil, nil
				},
			},
			notEmptyResult: true,
		},
		"receipts and transactions count mismatch": {
			backend: &backend{
				backendMock: newBackendMock(),
				chainConfig: &params.ChainConfig{},
				blockByHash: func(_ context.Context, hash common.Hash) (*types.Block, error) {
					return &types.Block{}, nil
				},
				getReceipts: func(ctx context.Context, hash common.Hash) (types.Receipts, error) {
					return types.Receipts{&types.Receipt{}}, nil
				},
			},
			expectedError: "receipts and transactions count mismatch",
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			api := NewTransactionAPI(testCase.backend, new(AddrLocker))

			blockReceipts, err := api.GetBlockReceipts(context.Background(), common.Hash{})

			if testCase.notEmptyResult {
				require.NotNil(t, blockReceipts)
				require.Len(t, blockReceipts, 0)
			} else {
				require.Nil(t, blockReceipts)
			}

			if testCase.expectedError != "" {
				require.ErrorContains(t, err, testCase.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSuccessfulCallsOfReceiptApis(t *testing.T) {
	signer := types.HomesteadSigner{}
	key, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")

	blockNumber := big.NewInt(1000)

	tx1, _ := types.SignNewTx(key, signer, &types.LegacyTx{
		To:       &common.Address{1},
		Nonce:    1,
		Value:    big.NewInt(1),
		Gas:      21000,
		GasPrice: big.NewInt(1000000000),
	})

	tx2, _ := types.SignNewTx(key, signer, &types.LegacyTx{
		Nonce:    2,
		Value:    big.NewInt(1),
		Gas:      100000,
		GasPrice: big.NewInt(1000000000),
	})

	receipt1 := &types.Receipt{Status: types.ReceiptStatusSuccessful, GasUsed: 21000, CumulativeGasUsed: 21000}
	receipt2 := &types.Receipt{
		Status:            types.ReceiptStatusSuccessful,
		GasUsed:           100000,
		CumulativeGasUsed: 121000,
		ContractAddress:   common.Address{2},
		Logs: []*types.Log{{
			Address:     common.Address{2},
			Data:        []byte{},
			BlockNumber: blockNumber.Uint64(),
			TxHash:      tx2.Hash(),
			TxIndex:     1,
		}},
	}

	receipts := types.Receipts{receipt1, receipt2}

	header := &types.Header{Number: blockNumber}
	block := types.NewBlock(header, []*types.Transaction{tx1, tx2}, nil, receipts, trie.NewStackTrie(nil))

	receipt1.BlockHash = block.Hash()
	receipt2.BlockHash = block.Hash()
	receipt2.Logs[0].BlockHash = block.Hash()
	receipt1.BlockNumber = block.Number()
	receipt2.BlockNumber = block.Number()

	b := &backend{
		backendMock: newBackendMock(),
		chainConfig: &params.ChainConfig{},
		blockByHash: func(_ context.Context, hash common.Hash) (*types.Block, error) {
			require.Equal(t, block.Hash(), hash)
			return block, nil
		},
		getReceipts: func(_ context.Context, hash common.Hash) (types.Receipts, error) {
			require.Equal(t, block.Hash(), hash)
			return receipts, nil
		},
		getTransaction: func(_ context.Context, hash common.Hash) (*types.Transaction, common.Hash, uint64, uint64, error) {
			switch hash {
			case tx1.Hash():
				return tx1, block.Hash(), block.NumberU64(), 0, nil
			case tx2.Hash():
				return tx2, block.Hash(), block.NumberU64(), 1, nil
			default:
				t.Error("unexpected transaction hash")
				return nil, common.Hash{}, 0, 0, nil
			}
		},
	}

	api := NewTransactionAPI(b, new(AddrLocker))

	expectedBloom := make([]byte, 256)

	expectedReceipt1Data := map[string]any{
		"blockHash":         "0xe279c849a2e438e68c2065dfb1a4d655f140ca7a1c629abb0a3d8f0c218507fa",
		"blockNumber":       "0x3e8",
		"contractAddress":   nil,
		"cumulativeGasUsed": "0x5208",
		"effectiveGasPrice": nil,
		"from":              "0x71562b71999873db5b286df957af199ec94617f7",
		"gasUsed":           "0x5208",
		"logs":              []any{},
		"logsBloom":         "0x" + hex.EncodeToString(expectedBloom),
		"status":            "0x1",
		"to":                "0x0100000000000000000000000000000000000000",
		"transactionHash":   "0xb3d436a588c817f3db0a20da78243b709c4eae25a60efdbd9aea61ea3b3bf5d9",
		"transactionIndex":  "0x0",
		"type":              "0x0",
	}

	expectedReceipt2Data := map[string]any{
		"blockHash":         "0xe279c849a2e438e68c2065dfb1a4d655f140ca7a1c629abb0a3d8f0c218507fa",
		"blockNumber":       "0x3e8",
		"contractAddress":   "0x0200000000000000000000000000000000000000",
		"cumulativeGasUsed": "0x1d8a8",
		"effectiveGasPrice": nil,
		"from":              "0x71562b71999873db5b286df957af199ec94617f7",
		"gasUsed":           "0x186a0",
		"logs": []any{
			map[string]any{
				"address":          "0x0200000000000000000000000000000000000000",
				"topics":           nil,
				"data":             "0x",
				"blockNumber":      "0x3e8",
				"transactionHash":  "0x4fc305da8c0d347c984c2fca38709b0d09f6f1978dd78cae6bc10f2d524f8650",
				"transactionIndex": "0x1",
				"blockHash":        "0xe279c849a2e438e68c2065dfb1a4d655f140ca7a1c629abb0a3d8f0c218507fa",
				"logIndex":         "0x0",
				"removed":          false,
			},
		},
		"logsBloom":        "0x" + hex.EncodeToString(expectedBloom),
		"status":           "0x1",
		"to":               nil,
		"transactionHash":  "0x4fc305da8c0d347c984c2fca38709b0d09f6f1978dd78cae6bc10f2d524f8650",
		"transactionIndex": "0x1",
		"type":             "0x0",
	}

	assertJsonEqual := func(expected, actual any) {
		expectedJson, err := json.Marshal(expected)
		require.NoError(t, err)
		actualJson, err := json.Marshal(actual)
		require.NoError(t, err)
		require.JSONEq(t, string(expectedJson), string(actualJson))
	}

	receipt1Data, err := api.GetTransactionReceipt(context.Background(), tx1.Hash())
	require.NoError(t, err)
	assertJsonEqual(expectedReceipt1Data, receipt1Data)

	receipt2Data, err := api.GetTransactionReceipt(context.Background(), tx2.Hash())
	require.NoError(t, err)
	assertJsonEqual(expectedReceipt2Data, receipt2Data)

	allReceiptsData, err := api.GetBlockReceipts(context.Background(), block.Hash())
	require.NoError(t, err)
	assertJsonEqual([]any{receipt1Data, receipt2Data}, allReceiptsData)
}
