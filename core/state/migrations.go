package state

import (
	"github.com/ethereum/go-ethereum/core/state/migrations"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

// Migration represents a single state migration.
type Migration interface {
	// Block returns block height for migration execution
	Block() *big.Int

	// Name returns migration name
	Name() string

	// Execute applies state changes to the specified state
	Execute(stateDB vm.StateDB)
}

// Migrations represents migrations lists mapped by block heights where these migrations should be executed.
type Migrations map[uint64][]Migration

// InitMigrations initializes state migrations list.
// If state should be changed in some unusual way that is not described in consensus rules,
// it can be done by writing a new state migration and registering it here,
// so migration will be applied at specific block and hardfork will happen.
func InitMigrations(config *params.ChainConfig) Migrations {
	output := make(Migrations)

	// All available migrations should be defined here
	availableMigrations := []Migration{
		migrations.NewMintContractMigration(config),
	}

	for _, migration := range availableMigrations {
		output.register(migration)
	}

	return output
}

func (m Migrations) register(migration Migration) {
	if migration.Block() == nil {
		return
	}

	block := migration.Block().Uint64()

	if existingMigrations, exists := m[block]; !exists {
		m[block] = []Migration{migration}
	} else {
		m[block] = append(existingMigrations, migration)
	}
}

// Execute finds state migrations for specified height and executes each of them for specified state.
func (m Migrations) Execute(height *big.Int, stateDB vm.StateDB, source string) {
	migrationsForBlock, exists := m[height.Uint64()]
	if !exists {
		return
	}

	for _, migration := range migrationsForBlock {
		log.Info("Executing state migration", "name", migration.Name(), "height", height.Uint64(), "source", source)
		migration.Execute(stateDB)
	}
}
