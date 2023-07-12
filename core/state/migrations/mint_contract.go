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
	block        *big.Int
	ownerAddress common.Address
	mintLimit    *big.Int
}

// NewMintContractMigration validates chain config
// and either creates a new migration instance, or returns a validation error.
func NewMintContractMigration(config *params.ChainConfig) (*MintContractMigration, error) {
	if config.MintContract == nil || config.MintContract.ActivationBlock == nil {
		return nil, nil
	}

	if config.MintContract.OwnerAddress == (common.Address{}) {
		return nil, errors.New("owner address is not specified or equals to zero address")
	}

	if config.MintContract.MintLimit == nil {
		return nil, errors.New("mint limit is not specified")
	}

	return &MintContractMigration{
		block:        config.MintContract.ActivationBlock,
		ownerAddress: config.MintContract.OwnerAddress,
		mintLimit:    (*big.Int)(config.MintContract.MintLimit),
	}, nil
}

func (m *MintContractMigration) Block() *big.Int {
	return m.block
}

func (m *MintContractMigration) Name() string {
	return "mint contract initialization"
}

func (m *MintContractMigration) Execute(stateDB vm.StateDB) {
	stateDB.SetCode(mint.Contract.Address, mint.Contract.Bytecode)
	stateDB.SetState(mint.Contract.Address, mint.Contract.StorageLayout.Owner, m.ownerAddress.Hash())
	stateDB.SetState(mint.Contract.Address, mint.Contract.StorageLayout.MintLimit, common.BigToHash(m.mintLimit))
}
