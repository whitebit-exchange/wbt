package core

import (
	"bytes"
	"crypto/ecdsa"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/mint"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"math"
	"math/big"
	"testing"
)

var ownerKey, _ = crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f291")
var ownerAddr = crypto.PubkeyToAddress(ownerKey.PublicKey)
var mintLimit = new(big.Int).Mul(big.NewInt(1000), big.NewInt(params.Ether))

func prepareStateDb() *state.StateDB {
	stateDb, _ := state.New(common.Hash{}, state.NewDatabase(rawdb.NewMemoryDatabase()), nil)

	stateDb.SetCode(mint.Contract.Address, mint.Contract.Bytecode)
	stateDb.SetState(mint.Contract.Address, mint.Contract.StorageLayout.Owner, ownerAddr.Hash())
	stateDb.SetState(mint.Contract.Address, mint.Contract.StorageLayout.MintLimit, common.BigToHash(mintLimit))
	stateDb.AddBalance(ownerAddr, big.NewInt(params.Ether))
	stateDb.Finalise(true)

	return stateDb
}

func TestIncorrectMintInstruction(t *testing.T) {
	sender2Key, _ := crypto.HexToECDSA("b71c71a67e1177ad4e901695e1b4b9ee17ae16c6668d313eac2f96dbcda3f292")
	sender2Addr := crypto.PubkeyToAddress(sender2Key.PublicKey)

	blockNum := big.NewInt(100)
	blockCtx := vm.BlockContext{
		CanTransfer: func(db vm.StateDB, address common.Address, b *big.Int) bool { return true },
		Transfer:    func(db vm.StateDB, address common.Address, address2 common.Address, b *big.Int) {},
		BlockNumber: blockNum,
	}

	chainConfig := params.AllEthashProtocolChanges
	chainConfig.LondonBlock = nil

	validMintAmount := new(big.Int).Mul(big.NewInt(500), big.NewInt(params.Ether))
	validData := bytes.Join([][]byte{
		common.BigToHash(validMintAmount).Bytes(),
		(common.Hash{}).Bytes(),
		{byte(1)},
	}, []byte{})

	testCases := []struct {
		signerKey     *ecdsa.PrivateKey
		tx            *types.LegacyTx
		expectedError string
		modifyStateDb func(stateDb *state.StateDB)
		usedGas       uint64
	}{
		{
			signerKey: ownerKey,
			tx: &types.LegacyTx{
				Nonce:    0,
				To:       &mint.Contract.Address,
				Value:    new(big.Int),
				Gas:      30000,
				Data:     bytes.Repeat([]byte{2}, 64),
				GasPrice: big.NewInt(params.GWei),
			},
			expectedError: "execution reverted",
			usedGas:       22194,
		},
		{
			signerKey: ownerKey,
			tx: &types.LegacyTx{
				Nonce:    0,
				To:       &mint.Contract.Address,
				Value:    new(big.Int),
				Gas:      130000,
				Data:     bytes.Repeat([]byte{2}, 65),
				GasPrice: big.NewInt(params.GWei),
			},
			expectedError: "invalid burn tx network in mint instruction",
			usedGas:       122040,
		},
		{
			signerKey: sender2Key,
			tx: &types.LegacyTx{
				Nonce:    0,
				To:       &mint.Contract.Address,
				Value:    new(big.Int),
				Gas:      130000,
				Data:     bytes.Repeat([]byte{0}, 65),
				GasPrice: big.NewInt(params.GWei),
			},
			expectedError: "transaction sender is not allowed to mint",
			modifyStateDb: func(stateDb *state.StateDB) {
				stateDb.AddBalance(sender2Addr, big.NewInt(params.Ether))
				stateDb.Finalise(true)
			},
			usedGas: 121260,
		},
		{
			signerKey: ownerKey,
			tx: &types.LegacyTx{
				Nonce:    0,
				To:       &mint.Contract.Address,
				Value:    new(big.Int),
				Gas:      130000,
				Data:     bytes.Repeat([]byte{1}, 65),
				GasPrice: big.NewInt(params.GWei),
			},
			expectedError: "mint amount exceeds mint limit",
			usedGas:       122040,
		},
		{
			signerKey: ownerKey,
			tx: &types.LegacyTx{
				Nonce:    0,
				To:       &mint.Contract.Address,
				Value:    new(big.Int),
				Gas:      130000,
				Data:     validData,
				GasPrice: big.NewInt(params.GWei),
			},
			expectedError: "execution reverted",
			modifyStateDb: func(stateDb *state.StateDB) {
				stateDb.SetCode(mint.Contract.Address, mint.Contract.Bytecode[:len(mint.Contract.Bytecode)-106])
				stateDb.Finalise(true)
			},
			usedGas: 21526,
		},
	}

	signer := types.HomesteadSigner{}

	for _, testCase := range testCases {
		t.Run(testCase.expectedError, func(t *testing.T) {
			stateDb := prepareStateDb()
			if testCase.modifyStateDb != nil {
				testCase.modifyStateDb(stateDb)
			}

			evm := vm.NewEVM(blockCtx, vm.TxContext{}, stateDb, chainConfig, vm.Config{NoBaseFee: true})

			tx, _ := types.SignNewTx(testCase.signerKey, signer, testCase.tx)
			message, _ := TransactionToMessage(tx, signer, nil)
			result, err := ApplyMessage(evm, message, new(GasPool).AddGas(math.MaxUint64))

			sender, _ := signer.Sender(tx)

			successful := assert.NoError(t, err) &&
				assert.ErrorContains(t, result.Err, testCase.expectedError) &&
				assert.Equal(t, testCase.usedGas, result.UsedGas) &&
				assert.Equal(t, uint64(1), stateDb.GetNonce(sender))

			if !successful {
				t.Error("test failed")
			}
		})
	}
}

func TestOutOfGas(t *testing.T) {
	stateDb := prepareStateDb()

	blockNum := big.NewInt(100)
	blockCtx := vm.BlockContext{
		CanTransfer: func(db vm.StateDB, address common.Address, b *big.Int) bool { return true },
		Transfer:    func(db vm.StateDB, address common.Address, address2 common.Address, b *big.Int) {},
		BlockNumber: blockNum,
	}

	chainConfig := params.AllCliqueProtocolChanges
	chainConfig.LondonBlock = nil

	mintAmount := new(big.Int).Mul(big.NewInt(500), big.NewInt(params.Ether))
	burnTxHash := common.HexToHash("0x621c759718a44e19ad04f8d133746b1043a2004f3fd68028cd28f1598388106e")
	burnTxNetwork := byte(0)

	data := bytes.Join([][]byte{common.BigToHash(mintAmount).Bytes(), burnTxHash.Bytes(), {burnTxNetwork}}, []byte{})

	evm := vm.NewEVM(blockCtx, vm.TxContext{}, stateDb, chainConfig, vm.Config{NoBaseFee: true})
	signer := types.HomesteadSigner{}

	tx, _ := types.SignNewTx(ownerKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &mint.Contract.Address,
		Value:    new(big.Int),
		Gas:      50000,
		Data:     data,
		GasPrice: big.NewInt(params.GWei),
	})

	message, _ := TransactionToMessage(tx, signer, nil)

	result, err := ApplyMessage(evm, message, new(GasPool).AddGas(math.MaxUint64))

	assert.NoError(t, err)
	assert.ErrorIs(t, result.Err, vm.ErrOutOfGas)
	assert.Equal(t, uint64(50000), result.UsedGas)
	assert.Equal(t, uint64(1), stateDb.GetNonce(ownerAddr))
}

func TestTransactionToCommonAddressWith65BytesData(t *testing.T) {
	stateDb := prepareStateDb()

	blockNum := big.NewInt(100)
	blockCtx := vm.BlockContext{
		CanTransfer: func(db vm.StateDB, address common.Address, b *big.Int) bool { return true },
		Transfer:    func(db vm.StateDB, address common.Address, address2 common.Address, b *big.Int) {},
		BlockNumber: blockNum,
	}

	feeCollectorAddress := common.HexToAddress("0x1000000000000000000000000000000000000000")

	chainConfig := &params.ChainConfig{
		ChainID:             big.NewInt(1337),
		HomesteadBlock:      big.NewInt(0),
		EIP150Block:         big.NewInt(0),
		EIP155Block:         big.NewInt(0),
		EIP158Block:         big.NewInt(0),
		ByzantiumBlock:      big.NewInt(0),
		ConstantinopleBlock: big.NewInt(0),
		PetersburgBlock:     big.NewInt(0),
		IstanbulBlock:       big.NewInt(0),
		MuirGlacierBlock:    big.NewInt(0),
		BerlinBlock:         big.NewInt(0),
		Clique:              &params.CliqueConfig{Period: 0, Epoch: 30000},
		FeeCollectorAddress: &feeCollectorAddress,
	}

	evm := vm.NewEVM(blockCtx, vm.TxContext{}, stateDb, chainConfig, vm.Config{NoBaseFee: true})
	signer := types.HomesteadSigner{}

	tx, _ := types.SignNewTx(ownerKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &common.Address{},
		Value:    new(big.Int),
		Gas:      130000,
		Data:     bytes.Repeat([]byte{1}, 65),
		GasPrice: big.NewInt(params.GWei),
	})

	message, _ := TransactionToMessage(tx, signer, nil)

	result, err := ApplyMessage(evm, message, new(GasPool).AddGas(math.MaxUint64))

	assert.NoError(t, err)
	assert.NoError(t, result.Err)
	assert.Equal(t, uint64(22040), result.UsedGas)

	feeAmount, _ := new(big.Int).SetString("22040000000000", 10)
	assert.Equal(t, feeAmount, stateDb.GetBalance(feeCollectorAddress))
	assert.Equal(t, new(big.Int), stateDb.GetBalance(evm.Context.Coinbase))

	// Fee = 22040 * 1 gwei
	// Previous balance = 1 ether
	// Expected balance = Previous balance - Fee = 0.99997796 ether
	expectedBalance, _ := new(big.Int).SetString("999977960000000000", 10)
	// Previous mint limit = 1000
	// Expected mint limit = Previous mint limit
	expectedMintLimit := common.BigToHash(mintLimit)

	assert.Equal(t, uint64(1), stateDb.GetNonce(ownerAddr))
	assert.Equal(t, expectedBalance, stateDb.GetBalance(ownerAddr))
	assert.Equal(t, expectedMintLimit, stateDb.GetState(mint.Contract.Address, mint.Contract.StorageLayout.MintLimit))
	assert.Len(t, stateDb.Logs(), 0)
}

func TestSuccessfulMint(t *testing.T) {
	stateDb := prepareStateDb()

	blockNum := big.NewInt(100)
	blockCtx := vm.BlockContext{
		CanTransfer: func(db vm.StateDB, address common.Address, b *big.Int) bool { return true },
		Transfer:    func(db vm.StateDB, address common.Address, address2 common.Address, b *big.Int) {},
		BlockNumber: blockNum,
	}

	chainConfig := params.AllCliqueProtocolChanges
	chainConfig.LondonBlock = nil

	mintAmount := new(big.Int).Mul(big.NewInt(500), big.NewInt(params.Ether))
	burnTxHash := common.HexToHash("0x621c759718a44e19ad04f8d133746b1043a2004f3fd68028cd28f1598388106e")
	burnTxNetwork := byte(0)

	data := bytes.Join([][]byte{common.BigToHash(mintAmount).Bytes(), burnTxHash.Bytes(), {burnTxNetwork}}, []byte{})

	evm := vm.NewEVM(blockCtx, vm.TxContext{}, stateDb, chainConfig, vm.Config{NoBaseFee: true})
	signer := types.HomesteadSigner{}

	tx, _ := types.SignNewTx(ownerKey, signer, &types.LegacyTx{
		Nonce:    0,
		To:       &mint.Contract.Address,
		Value:    new(big.Int),
		Gas:      130000,
		Data:     data,
		GasPrice: big.NewInt(params.GWei),
	})

	message, _ := TransactionToMessage(tx, signer, nil)

	result, err := ApplyMessage(evm, message, new(GasPool).AddGas(math.MaxUint64))

	assert.NoError(t, err)
	assert.NoError(t, result.Err)
	assert.Equal(t, uint64(121716), result.UsedGas)

	// Fee = 121716 * 1 gwei
	feeAmount, _ := new(big.Int).SetString("121716000000000", 10)
	assert.Equal(t, feeAmount, stateDb.GetBalance(evm.Context.Coinbase))

	// Fee = 121716 * 1 gwei
	// Mint amount = 500 * 1 ether
	// Previous balance = 1 ether
	// Expected balance = Previous balance - Fee + Mint amount = 500.999878284 ether
	expectedBalance, _ := new(big.Int).SetString("500999878284000000000", 10)
	// Previous mint limit = 1000
	// Expected mint limit = Previous mint limit - Mint amount
	expectedMintLimit := common.BigToHash(new(big.Int).Mul(big.NewInt(500), big.NewInt(params.Ether)))

	assert.Equal(t, uint64(1), stateDb.GetNonce(ownerAddr))
	assert.Equal(t, expectedBalance, stateDb.GetBalance(ownerAddr))
	assert.Equal(t, expectedMintLimit, stateDb.GetState(mint.Contract.Address, mint.Contract.StorageLayout.MintLimit))

	assert.Len(t, stateDb.Logs(), 1)
	assert.Equal(t, &types.Log{
		Address: mint.Contract.Address,
		Topics:  []common.Hash{common.HexToHash("0d9811f14a9cfa628d4819902adcdd4ff09f73ac9c2628280058dc2146fa247d")},
		Data: bytes.Join([][]byte{
			common.BigToHash(mintAmount).Bytes(),
			burnTxHash.Bytes(),
			common.BytesToHash([]byte{burnTxNetwork}).Bytes(),
		}, []byte{}),
		BlockNumber: blockNum.Uint64(),
	}, stateDb.Logs()[0])
}
