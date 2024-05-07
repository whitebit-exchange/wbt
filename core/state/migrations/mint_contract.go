package migrations

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/mint"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
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

// NewMintContractMigration creates a new migration instance.
// If mint contract config is specified in chain config and is valid, an initialised migration will be returned.
// If mint contract config is not specified, empty migration will be returned, so it will not be registered.
// Invalid mint contract config will cause a critical error.
func NewMintContractMigration(config *params.ChainConfig) *MintContractMigration {
	if config.MintContract == nil || config.MintContract.ActivationBlock == nil {
		return &MintContractMigration{}
	}

	if config.MintContract.OwnerAddress == (common.Address{}) {
		log.Crit("invalid mint contract config: owner address is not specified or equals to zero address")
	}

	if config.MintContract.MintLimit == nil {
		log.Crit("invalid mint contract config: mint limit is not specified")
	}

	return &MintContractMigration{
		block:        config.MintContract.ActivationBlock,
		ownerAddress: config.MintContract.OwnerAddress,
		mintLimit:    (*big.Int)(config.MintContract.MintLimit),
	}
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
