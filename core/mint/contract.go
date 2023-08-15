package mint

import (
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"strings"
)

const (
	BurnNetworkEthereum byte = iota
	BurnNetworkTron
)

const EventName = "Mint"
const eventAbiString = `[{
	"anonymous": false,
	"inputs": [
		{"indexed": false, "internalType": "uint256", "name": "amount", "type": "uint256"},
		{"indexed": false, "internalType": "bytes32", "name": "burnTxHash", "type": "bytes32"},
		{"indexed": false, "internalType": "uint8", "name": "burnTxNetwork", "type": "uint8"}
	],
	"name": "Mint",
	"type": "event"
}]`

var EventAbi, _ = abi.JSON(strings.NewReader(eventAbiString))

type storageLayout struct {
	Owner     common.Hash
	MintLimit common.Hash
}

// Contract is an object representing predefined mint contract params.
// Specified bytecode is a result of MintState contract compilation (see core/mint/contract/MintState.sol).
// Standard JSON input that is used for compilation is located near contract and can be used to reproduce the bytecode.
// BytecodeHash is a result of crypto.Keccak256Hash(bytecode).
var Contract = struct {
	Address       common.Address
	Bytecode      []byte
	BytecodeHash  common.Hash
	StorageLayout storageLayout
}{
	Address:      common.HexToAddress(contractAddress),
	Bytecode:     common.Hex2Bytes(contractBytecode),
	BytecodeHash: common.HexToHash(contractBytecodeHash),
	StorageLayout: storageLayout{
		Owner:     common.Hash{},
		MintLimit: common.BytesToHash([]byte{1}),
	},
}
