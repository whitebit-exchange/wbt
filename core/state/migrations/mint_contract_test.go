package migrations

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/params"
	"github.com/stretchr/testify/assert"
	"math/big"
	"testing"
)

func TestInvalidConfig(t *testing.T) {
	testCases := []struct {
		config          *params.MintContractConfig
		expectedMessage string
	}{
		{&params.MintContractConfig{}, "owner address is not specified or equals to zero address"},
		{&params.MintContractConfig{OwnerAddress: common.BytesToAddress([]byte{1})}, "mint limit is not specified"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.expectedMessage, func(t *testing.T) {
			testCase.config.ActivationBlock = big.NewInt(1)
			_, err := NewMintContractMigration(&params.ChainConfig{MintContract: testCase.config})
			assert.ErrorContains(t, err, testCase.expectedMessage)
		})
	}
}
