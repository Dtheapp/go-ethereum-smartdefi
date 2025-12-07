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

// DecodeCreateTokenInput decodes the createAssetBackedToken input (parameters only, no method ID)
func DecodeCreateTokenInput(input []byte) (TokenConfig, error) {
	var config TokenConfig
	method := precompileABI.Methods["createAssetBackedToken"]
	values, err := method.Inputs.Unpack(input)
	if err != nil {
		return config, err
	}
	// Unpack the tuple into the struct
	return config, method.Inputs.Copy(&config, values)
}

// EncodeGetBacking encodes the getBacking call
func EncodeGetBacking(token common.Address, amount *big.Int) ([]byte, error) {
	return precompileABI.Pack("getBacking", token, amount)
}

// DecodeGetBackingInput decodes the getBacking input (parameters only, no method ID)
func DecodeGetBackingInput(input []byte) (common.Address, *big.Int, error) {
	method := precompileABI.Methods["getBacking"]
	values, err := method.Inputs.Unpack(input)
	if err != nil {
		return common.Address{}, nil, err
	}
	if len(values) < 2 {
		return common.Address{}, nil, errors.New("insufficient values")
	}
	token, ok1 := values[0].(common.Address)
	amount, ok2 := values[1].(*big.Int)
	if !ok1 || !ok2 {
		return common.Address{}, nil, errors.New("type assertion failed")
	}
	return token, amount, nil
}

// EncodeBurnAndRecover encodes the burnAndRecover call
func EncodeBurnAndRecover(token common.Address, amount *big.Int) ([]byte, error) {
	return precompileABI.Pack("burnAndRecover", token, amount)
}

// DecodeBurnAndRecoverInput decodes the burnAndRecover input (parameters only, no method ID)
func DecodeBurnAndRecoverInput(input []byte) (common.Address, *big.Int, error) {
	method := precompileABI.Methods["burnAndRecover"]
	values, err := method.Inputs.Unpack(input)
	if err != nil {
		return common.Address{}, nil, err
	}
	if len(values) < 2 {
		return common.Address{}, nil, errors.New("insufficient values")
	}
	token, ok1 := values[0].(common.Address)
	amount, ok2 := values[1].(*big.Int)
	if !ok1 || !ok2 {
		return common.Address{}, nil, errors.New("type assertion failed")
	}
	return token, amount, nil
}

// EncodeGetFloorPrice encodes the getFloorPrice call
func EncodeGetFloorPrice(token common.Address) ([]byte, error) {
	return precompileABI.Pack("getFloorPrice", token)
}

// DecodeGetFloorPriceInput decodes the getFloorPrice input (parameters only, no method ID)
func DecodeGetFloorPriceInput(input []byte) (common.Address, error) {
	method := precompileABI.Methods["getFloorPrice"]
	values, err := method.Inputs.Unpack(input)
	if err != nil {
		return common.Address{}, err
	}
	if len(values) < 1 {
		return common.Address{}, errors.New("insufficient values")
	}
	token, ok := values[0].(common.Address)
	if !ok {
		return common.Address{}, errors.New("type assertion failed")
	}
	return token, nil
}

// EncodeOutput encodes function output
func EncodeOutput(method string, output interface{}) ([]byte, error) {
	methodObj, ok := precompileABI.Methods[method]
	if !ok {
		return nil, errors.New("method not found")
	}
	return methodObj.Outputs.Pack(output)
}

