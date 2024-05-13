package state

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/mint"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state/migrations"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
)

func TestInitEmptyMigrationsList(t *testing.T) {
	migrationsList := InitMigrations(&params.ChainConfig{})
	assert.Len(t, migrationsList, 0)
}

func TestInitMigrationsAndApplyOnRandomBlock(t *testing.T) {
	chainConfig := &params.ChainConfig{
		MintContract: &params.MintContractConfig{
			ActivationBlock: big.NewInt(100000),
			OwnerAddress:    common.HexToAddress("0x1000000000000000000000000000000000000000"),
			MintLimit:       math.NewHexOrDecimal256(1000000000000000000),
		},
	}

	migrationsList := InitMigrations(chainConfig)
	assert.Len(t, migrationsList, 1)
	assert.Len(t, migrationsList[100000], 1)

	migrationsList.register(migrationsList[100000][0])
	assert.Len(t, migrationsList, 1)
	assert.Len(t, migrationsList[100000], 2)

	_, casted := migrationsList[100000][0].(*migrations.MintContractMigration)
	assert.True(t, casted)

	state, _ := New(common.Hash{}, NewDatabase(rawdb.NewMemoryDatabase()), nil)

	migrationsList.Execute(big.NewInt(100001), state, "tests")
	assert.Len(t, state.journal.entries, 0)
}

func TestApplyMintContractMigration(t *testing.T) {
	mintOwner := common.HexToAddress("0x1000000000000000000000000000000000000000")
	mintLimit := big.NewInt(1000000000000000000)
	activationBlock := big.NewInt(100000)

	migrationsList := InitMigrations(&params.ChainConfig{
		MintContract: &params.MintContractConfig{
			ActivationBlock: activationBlock,
			OwnerAddress:    mintOwner,
			MintLimit:       (*math.HexOrDecimal256)(mintLimit),
		},
	})

	state, _ := New(common.Hash{}, NewDatabase(rawdb.NewMemoryDatabase()), nil)

	migrationsList.Execute(activationBlock, state, "tests")

	contractStateObject := state.getStateObject(mint.Contract.Address)
	assert.Equal(t, new(big.Int), contractStateObject.Balance())
	assert.Equal(t, uint64(0), contractStateObject.Nonce())
	assert.Equal(t, mint.Contract.BytecodeHash.Bytes(), contractStateObject.data.CodeHash)

	assert.Equal(t, map[common.Hash]common.Hash{
		mint.Contract.StorageLayout.Owner:     {},
		mint.Contract.StorageLayout.MintLimit: {},
	}, map[common.Hash]common.Hash(contractStateObject.originStorage))

	assert.Equal(t, map[common.Hash]common.Hash{
		mint.Contract.StorageLayout.Owner:     common.BytesToHash(mintOwner[:]),
		mint.Contract.StorageLayout.MintLimit: common.BigToHash(mintLimit),
	}, map[common.Hash]common.Hash(contractStateObject.dirtyStorage))

	assert.Len(t, state.journal.entries, 4)

	createObjectChangeEntry, casted := state.journal.entries[0].(createObjectChange)
	assert.True(t, casted)
	assert.Equal(t, mint.Contract.Address, *createObjectChangeEntry.account)

	codeChangeEntry, casted := state.journal.entries[1].(codeChange)
	assert.True(t, casted)
	assert.Equal(t, mint.Contract.Address, *codeChangeEntry.account)
	assert.Equal(t, []byte(nil), codeChangeEntry.prevcode)

	ownerChangeEntry, casted := state.journal.entries[2].(storageChange)
	assert.True(t, casted)
	assert.Equal(t, mint.Contract.Address, *ownerChangeEntry.account)
	assert.Equal(t, mint.Contract.StorageLayout.Owner, ownerChangeEntry.key)
	assert.Equal(t, common.Hash{}, ownerChangeEntry.prevalue)

	mintLimitChangeEntry, casted := state.journal.entries[3].(storageChange)
	assert.True(t, casted)
	assert.Equal(t, mint.Contract.Address, *mintLimitChangeEntry.account)
	assert.Equal(t, mint.Contract.StorageLayout.MintLimit, mintLimitChangeEntry.key)
	assert.Equal(t, common.Hash{}, mintLimitChangeEntry.prevalue)
}

func TestApplyCassiopeiaMigration(t *testing.T) {
	soulDropAddress := common.HexToAddress("0x1000000000000000000000000000000000000001")
	holdAmountAddress := common.HexToAddress("0x1000000000000000000000000000000000000002")
	activationBlock := big.NewInt(100000)

	migrationsList := InitMigrations(&params.ChainConfig{
		SystemContracts: &params.SystemContracts{
			SoulDrop:   soulDropAddress,
			HoldAmount: holdAmountAddress,
		},
		CassiopeiaBlock: activationBlock,
	})

	state, _ := New(common.Hash{}, NewDatabase(rawdb.NewMemoryDatabase()), nil)

	migrationsList.Execute(activationBlock, state, "tests")

	holdAmountStateObject := state.getStateObject(holdAmountAddress)
	assert.Equal(t, new(big.Int), holdAmountStateObject.Balance())
	assert.Equal(t, uint64(0), holdAmountStateObject.Nonce())
	assert.Equal(t, common.Hex2Bytes("cd59ceb7b7378c248ca0e14afafb7c2bd1aac6a437ea3e602f9c4b2fd94bb565"), holdAmountStateObject.data.CodeHash)

	soulDropStateObject := state.getStateObject(soulDropAddress)
	assert.Equal(t, new(big.Int), soulDropStateObject.Balance())
	assert.Equal(t, uint64(0), soulDropStateObject.Nonce())

	assert.Equal(t, map[common.Hash]common.Hash{
		common.HexToHash("0xcbc4e5fb02c3d1de23a9f1e014b4d2ee5aeaea9505df5e855c9210bf472495af"): {},
		common.HexToHash("0x83ec6a1f0257b830b5e016457c9cf1435391bf56cc98f369a58a54fe93772465"): {},
		common.HexToHash("0x405aad32e1adbac89bb7f176e338b8fc6e994ca210c9bb7bdca249b465942250"): {},
		common.HexToHash("0xc69056f16cbaa3c616b828e333ab7d3a32310765507f8f58359e99ebb7a885f3"): {},
		common.HexToHash("0xf2c49132ed1cee2a7e75bde50d332a2f81f1d01e5456d8a19d1df09bd561dbd2"): {},
		common.HexToHash("0x85aaa47b6dc46495bb8824fad4583769726fea36efd831a35556690b830a8fbe"): {},
		common.HexToHash("0x8a8dc4e5242ea8b1ab1d60606dae757e6c2cca9f92a2cced9f72c19960bcb458"): {},
		common.HexToHash("0x9dcb9783ba5cd0b54745f65f4f918525e461e91888c334e5342cb380ac558d53"): {},
		common.HexToHash("0x2d72af3c1b2b2956e6f694fb741556d5ca9524373974378cdbec16afa8b84164"): {},
		common.HexToHash("0xd56a60595ebefebed7f22dcee6c2acc61b06cf8c68e84c88677840365d1ff92b"): {},
	}, map[common.Hash]common.Hash(soulDropStateObject.originStorage))

	expectedSoulDropChanges := map[common.Hash]common.Hash{
		common.HexToHash("0xcbc4e5fb02c3d1de23a9f1e014b4d2ee5aeaea9505df5e855c9210bf472495af"): common.HexToHash("0x4e8329a9"),
		common.HexToHash("0x83ec6a1f0257b830b5e016457c9cf1435391bf56cc98f369a58a54fe93772465"): common.HexToHash("0x4e86b821"),
		common.HexToHash("0x405aad32e1adbac89bb7f176e338b8fc6e994ca210c9bb7bdca249b465942250"): common.HexToHash("0x4e91f31c"),
		common.HexToHash("0xc69056f16cbaa3c616b828e333ab7d3a32310765507f8f58359e99ebb7a885f3"): common.HexToHash("0x4e9f0cf6"),
		common.HexToHash("0xf2c49132ed1cee2a7e75bde50d332a2f81f1d01e5456d8a19d1df09bd561dbd2"): common.HexToHash("0x4ec0bbb8"),
		common.HexToHash("0x85aaa47b6dc46495bb8824fad4583769726fea36efd831a35556690b830a8fbe"): common.HexToHash("0x4f14e793"),
		common.HexToHash("0x8a8dc4e5242ea8b1ab1d60606dae757e6c2cca9f92a2cced9f72c19960bcb458"): common.HexToHash("0x4f97bcd6"),
		common.HexToHash("0x9dcb9783ba5cd0b54745f65f4f918525e461e91888c334e5342cb380ac558d53"): common.HexToHash("0x50526e56"),
		common.HexToHash("0x2d72af3c1b2b2956e6f694fb741556d5ca9524373974378cdbec16afa8b84164"): common.HexToHash("0x55c2c94b"),
		common.HexToHash("0xd56a60595ebefebed7f22dcee6c2acc61b06cf8c68e84c88677840365d1ff92b"): common.HexToHash("0x6401c8b1"),
	}

	assert.Equal(t, expectedSoulDropChanges, map[common.Hash]common.Hash(soulDropStateObject.dirtyStorage))

	assert.Len(t, state.journal.entries, 13)

	createObjectChangeEntry, casted := state.journal.entries[0].(createObjectChange)
	assert.True(t, casted)
	assert.Equal(t, holdAmountAddress, *createObjectChangeEntry.account)

	codeChangeEntry, casted := state.journal.entries[1].(codeChange)
	assert.True(t, casted)
	assert.Equal(t, holdAmountAddress, *codeChangeEntry.account)
	assert.Equal(t, []byte(nil), codeChangeEntry.prevcode)

	createObjectChangeEntry, casted = state.journal.entries[2].(createObjectChange)
	assert.True(t, casted)
	assert.Equal(t, soulDropAddress, *createObjectChangeEntry.account)

	for i := 3; i < 13; i++ {
		storageChangeEntry, casted := state.journal.entries[i].(storageChange)
		assert.True(t, casted)
		assert.Equal(t, soulDropAddress, *storageChangeEntry.account)

		_, exists := expectedSoulDropChanges[storageChangeEntry.key]
		assert.True(t, exists)
		assert.Equal(t, common.Hash{}, storageChangeEntry.prevalue)
		delete(expectedSoulDropChanges, storageChangeEntry.key)
	}
}
