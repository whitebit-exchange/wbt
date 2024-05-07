package state

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/core/mint"
	"github.com/ethereum/go-ethereum/core/rawdb"
	"github.com/ethereum/go-ethereum/core/state/migrations"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
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
