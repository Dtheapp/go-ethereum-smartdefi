// Package assetbacking - ABI definitions and encoding/decoding
package assetbacking

import (
	"bytes"
	"errors"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

// ABI definition for the asset backing precompile
const PrecompileABI = `[
	{
		"inputs": [{
			"components": [
				{"name": "name", "type": "string"},
				{"name": "symbol", "type": "string"},
				{"name": "totalSupply", "type": "uint256"},
				{"name": "backingAsset", "type": "address", "description": "Must be address(0) for Smart coin (native coin) - only option"},
				{"name": "initialBacking", "type": "uint256"},
				{"name": "fees", "type": "uint256[12]"},
				{"name": "onlySB", "type": "bool"},
				{"name": "owner", "type": "address"},
				{"name": "enableLGE", "type": "bool"}
			],
			"name": "config",
			"type": "tuple"
		}],
		"name": "createAssetBackedToken",
		"outputs": [{"name": "tokenAddress", "type": "address"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "getBacking",
		"outputs": [{"name": "backingAmount", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	},
	{
		"inputs": [
			{"name": "token", "type": "address"},
			{"name": "amount", "type": "uint256"}
		],
		"name": "burnAndRecover",
		"outputs": [{"name": "recoveredAmount", "type": "uint256"}],
		"stateMutability": "nonpayable",
		"type": "function"
	},
	{
		"inputs": [{"name": "token", "type": "address"}],
		"name": "getFloorPrice",
		"outputs": [{"name": "floorPrice", "type": "uint256"}],
		"stateMutability": "view",
		"type": "function"
	}
]`

var (
	precompileABI abi.ABI
)

func init() {
	var err error
	precompileABI, err = abi.JSON(bytes.NewReader([]byte(PrecompileABI)))
	if err != nil {
		panic(err)
	}
}

// EncodeCreateToken encodes the createAssetBackedToken call
func EncodeCreateToken(config TokenConfig) ([]byte, error) {
	return precompileABI.Pack("createAssetBackedToken", config)
}

// DecodeCreateTokenInput decodes the createAssetBackedToken input
func DecodeCreateTokenInput(input []byte) (TokenConfig, error) {
	var config TokenConfig
	err := precompileABI.UnpackIntoInterface(&config, "createAssetBackedToken", input)
	return config, err
}

// EncodeGetBacking encodes the getBacking call
func EncodeGetBacking(token common.Address, amount *big.Int) ([]byte, error) {
	return precompileABI.Pack("getBacking", token, amount)
}

// DecodeGetBackingInput decodes the getBacking input
func DecodeGetBackingInput(input []byte) (common.Address, *big.Int, error) {
	var results struct {
		Token  common.Address
		Amount *big.Int
	}
	err := precompileABI.UnpackIntoInterface(&results, "getBacking", input)
	return results.Token, results.Amount, err
}

// EncodeBurnAndRecover encodes the burnAndRecover call
func EncodeBurnAndRecover(token common.Address, amount *big.Int) ([]byte, error) {
	return precompileABI.Pack("burnAndRecover", token, amount)
}

// DecodeBurnAndRecoverInput decodes the burnAndRecover input
func DecodeBurnAndRecoverInput(input []byte) (common.Address, *big.Int, error) {
	var results struct {
		Token  common.Address
		Amount *big.Int
	}
	err := precompileABI.UnpackIntoInterface(&results, "burnAndRecover", input)
	return results.Token, results.Amount, err
}

// EncodeGetFloorPrice encodes the getFloorPrice call
func EncodeGetFloorPrice(token common.Address) ([]byte, error) {
	return precompileABI.Pack("getFloorPrice", token)
}

// DecodeGetFloorPriceInput decodes the getFloorPrice input
func DecodeGetFloorPriceInput(input []byte) (common.Address, error) {
	var token common.Address
	err := precompileABI.UnpackIntoInterface(&token, "getFloorPrice", input)
	return token, err
}

// EncodeOutput encodes function output
func EncodeOutput(method string, output interface{}) ([]byte, error) {
	methodObj, ok := precompileABI.Methods[method]
	if !ok {
		return nil, errors.New("method not found")
	}
	return methodObj.Outputs.Pack(output)
}

