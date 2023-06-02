package mint

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"math/big"
)

// state is a facade for accessing mint contract properties at the specified EVM state
type state struct {
	stateDB vm.StateDB
}

// stateFromDb returns a new facade instance if mint contract code hash equals to a predefined value
func stateFromDb(stateDB vm.StateDB) *state {
	if stateDB.GetCodeHash(Contract.Address) != Contract.BytecodeHash {
		return nil
	}

	return &state{stateDB: stateDB}
}

func (s *state) getOwner() common.Address {
	return common.BytesToAddress(s.getState(Contract.StorageLayout.Owner).Bytes())
}

func (s *state) getMintLimit() *big.Int {
	return s.getState(Contract.StorageLayout.MintLimit).Big()
}

func (s *state) getState(slotId common.Hash) common.Hash {
	return s.stateDB.GetState(Contract.Address, slotId)
}
