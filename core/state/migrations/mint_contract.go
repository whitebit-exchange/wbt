package migrations

import (
	"errors"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/mint"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
	"math/big"
)

// MintContractMigration initializes mint contract on a predefined address
// with a state retrieved from chain config.
type MintContractMigration struct {
	ownerAddress common.Address
	mintLimit    *big.Int
}

var MintContractMigrationSingleton = &MintContractMigration{}

// Init validates chain config
// and either sets up singleton, or returns a validation error.
func (m *MintContractMigration) Init(config *params.ChainConfig) (uint64, error) {
	if config.MintContract != nil && config.MintContract.ActivationBlock != nil {
		if config.MintContract.OwnerAddress == (common.Address{}) {
			return 0, errors.New("owner address is not specified or equals to zero address")
		}

		if config.MintContract.MintLimit == nil {
			return 0, errors.New("mint limit is not specified")
		}

		m.ownerAddress = config.MintContract.OwnerAddress
		m.mintLimit = (*big.Int)(config.MintContract.MintLimit)
		return config.MintContract.ActivationBlock.Uint64(), nil
	}

	return 0, errors.New("mint contract is not specified")
}

func (m *MintContractMigration) Name() string {
	return "mint contract initialization"
}

func (m *MintContractMigration) Execute(stateDB vm.StateDB) {
	stateDB.SetCode(mint.Contract.Address, mint.Contract.Bytecode)
	stateDB.SetState(mint.Contract.Address, mint.Contract.StorageLayout.Owner, m.ownerAddress.Hash())
	stateDB.SetState(mint.Contract.Address, mint.Contract.StorageLayout.MintLimit, common.BigToHash(m.mintLimit))
}
