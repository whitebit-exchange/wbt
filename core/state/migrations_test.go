package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state/migrations"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

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

	migrationsList.append(100000, migrationsList[100000][0])
	assert.Len(t, migrationsList, 1)
	assert.Len(t, migrationsList[100000], 2)

	_, casted := migrationsList[100000][0].(*migrations.MintContractMigration)
	assert.True(t, casted)

	state, _ := New(common.Hash{}, NewDatabase(rawdb.NewMemoryDatabase()), nil)

	migrationsList.Execute(big.NewInt(100001), state, "tests")
	assert.Len(t, state.journal.entries, 0)
}

func TestApplyMintContractMigration(t *testing.T) {
	migrationsList := InitMigrations(&params.ChainConfig{
		MintContract: &params.MintContractConfig{
			ActivationBlock: big.NewInt(100000),
			OwnerAddress:    common.HexToAddress("0x1000000000000000000000000000000000000000"),
			MintLimit:       math.NewHexOrDecimal256(1000000000000000000),
		},
	})

	state, _ := New(common.Hash{}, NewDatabase(rawdb.NewMemoryDatabase()), nil)

	migrationsList.Execute(big.NewInt(100000), state, "tests")

	expectedOwner := common.HexToHash("0x0000000000000000000000001000000000000000000000000000000000000000")
	expectedMintLimit := common.HexToHash("0x0000000000000000000000000000000000000000000000000de0b6b3a7640000")
	expectedBytecodeHash := common.HexToHash("0x094c2e08801f704f590fc847deec6076c880cdc6062f87c32614a4ff213fdf9c")

	contractStateObject := state.getStateObject(common.BytesToAddress([]byte{0x10, 0x00}))
	assert.Equal(t, new(big.Int), contractStateObject.Balance())
	assert.Equal(t, uint64(0), contractStateObject.Nonce())
	assert.Equal(t, expectedBytecodeHash.Bytes(), contractStateObject.data.CodeHash)

	assert.Equal(t, map[common.Hash]common.Hash{
		common.Hash{}:                 {},
		common.BytesToHash([]byte{1}): {},
	}, map[common.Hash]common.Hash(contractStateObject.originStorage))

	assert.Equal(t, map[common.Hash]common.Hash{
		common.Hash{}:                   expectedOwner,
		common.BytesToHash([]byte{0x1}): expectedMintLimit,
	}, map[common.Hash]common.Hash(contractStateObject.dirtyStorage))

	assert.Len(t, state.journal.entries, 4)

	createObjectChangeEntry, casted := state.journal.entries[0].(createObjectChange)
	assert.True(t, casted)
	assert.Equal(t, common.BytesToAddress([]byte{0x10, 0x00}), *createObjectChangeEntry.account)

	codeChangeEntry, casted := state.journal.entries[1].(codeChange)
	assert.True(t, casted)
	assert.Equal(t, common.BytesToAddress([]byte{0x10, 0x00}), *codeChangeEntry.account)
	assert.Equal(t, []byte(nil), codeChangeEntry.prevcode)

	ownerChangeEntry, casted := state.journal.entries[2].(storageChange)
	assert.True(t, casted)
	assert.Equal(t, common.BytesToAddress([]byte{0x10, 0x00}), *ownerChangeEntry.account)
	assert.Equal(t, common.Hash{}, ownerChangeEntry.key)
	assert.Equal(t, common.Hash{}, ownerChangeEntry.prevalue)

	mintLimitChangeEntry, casted := state.journal.entries[3].(storageChange)
	assert.True(t, casted)
	assert.Equal(t, common.BytesToAddress([]byte{0x10, 0x00}), *mintLimitChangeEntry.account)
	assert.Equal(t, common.BytesToHash([]byte{0x1}), mintLimitChangeEntry.key)
	assert.Equal(t, common.Hash{}, mintLimitChangeEntry.prevalue)
}
