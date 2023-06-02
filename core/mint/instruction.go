package mint

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/log"
	"math/big"
)

const (
	networkEthereum byte = iota
	networkTron
)

type Instruction struct {
	sender        common.Address
	amount        *big.Int
	nextLimit     *big.Int
	burnTxHash    common.Hash
	burnTxNetwork byte
	data          []byte
}

// ParseInstruction tries to find a mint instruction in transaction input data.
//
// Transaction is considered as a mint instruction if:
//	- tx data contains exactly 65 bytes (32 bytes - mint amount, 32 bytes - burn tx hash, 1 byte - burn tx network)
//	- receiver is a predefined mint contract address
//	- mint contract address contains expected mint contract code
//	- transaction sender equals to mint contract owner
//	- specified mint amount would not exceed mint limit
// Burn tx hash/network are used as a reference to a corresponding burn transaction in original network.
// These fields are only be applied to a Mint event.
// If any condition fails, nil will be returned instead of instruction, so transaction would be handled by the EVM.
// Otherwise, parsed instruction will be applied on the current state (see core/state_transition.go)
func ParseInstruction(evm *vm.EVM, sender, receiver common.Address, data []byte) *Instruction {
	if len(data) != 65 || receiver != Contract.Address {
		return nil
	}

	amountBytes, burnTxHashBytes, burnTxNetworkBytes := data[:32], data[32:64], data[64]
	if burnTxNetworkBytes != networkEthereum && burnTxNetworkBytes != networkTron {
		log.Warn("invalid burn tx network in mint instruction", "network", burnTxNetworkBytes)
		return nil
	}

	mintState := stateFromDb(evm.StateDB)
	if mintState == nil {
		log.Warn("mint contract not found in current state")
		return nil
	}

	contractOwner := mintState.getOwner()
	if sender != contractOwner {
		log.Warn("transaction sender is not allowed to mint", "sender", sender.Hex(), "owner", contractOwner.Hex())
		return nil
	}

	amount := new(big.Int).SetBytes(amountBytes)
	mintLimit := mintState.getMintLimit()
	if amount.Cmp(mintLimit) == 1 {
		log.Warn("mint amount exceeds mint limit", "amount", amount.String(), "limit", mintLimit.String())
		return nil
	}

	return &Instruction{
		sender:        sender,
		amount:        amount,
		nextLimit:     new(big.Int).Sub(mintLimit, amount),
		burnTxHash:    common.BytesToHash(burnTxHashBytes),
		burnTxNetwork: burnTxNetworkBytes,
		data:          data,
	}
}

// Apply modifies states of both mint contract and tx sender (mint contract owner) and produces Mint event.
func (i *Instruction) Apply(evm *vm.EVM, gas uint64) {
	evm.StateDB.SetState(Contract.Address, Contract.StorageLayout.MintLimit, common.BytesToHash(i.nextLimit.Bytes()))
	evm.StateDB.SetNonce(i.sender, evm.StateDB.GetNonce(i.sender)+1)
	evm.StateDB.AddBalance(i.sender, i.amount)

	event := mintEventAbi.Events[mintEventName]
	data, err := event.Inputs.Pack(i.amount, i.burnTxHash, i.burnTxNetwork)
	if err != nil {
		log.Crit("failed to pack Mint event data", "error", err)
	}

	evm.StateDB.AddLog(&types.Log{
		Address:     Contract.Address,
		Topics:      []common.Hash{event.ID},
		Data:        data,
		BlockNumber: evm.Context.BlockNumber.Uint64(),
	})

	if evm.Config.Debug {
		evm.Config.Tracer.CaptureStart(evm, i.sender, Contract.Address, false, i.data, gas, nil)
		evm.Config.Tracer.CaptureEnd(nil, 0, 0, nil)
	}
}
