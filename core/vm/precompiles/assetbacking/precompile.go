// Package assetbacking implements the native asset-backed token precompile
// Address: 0x0000000000000000000000000000000000000100
package assetbacking

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	
	"github.com/ethereum/go-ethereum/core/state/backingpool"
)

const (
	// PrecompileAddress is the address where this precompile is deployed
	PrecompileAddress = "0x0000000000000000000000000000000000000100"
	
	// Gas costs
	GasCreateToken      = 100000  // Base cost for token creation
	GasGetBacking       = 5000    // Cost for getting backing info
	GasBurnAndRecover   = 30000   // Cost for burn and recover
	GasPerByte          = 200     // Additional gas per byte of data
)

var (
	// PrecompileAddressBytes is the address as bytes
	PrecompileAddressBytes = common.HexToAddress(PrecompileAddress)
	
	// Method IDs (first 4 bytes of keccak256 hash of function signature)
	MethodIDCreateToken    = crypto.Keccak256([]byte("createAssetBackedToken((string,string,uint256,address,uint256,uint256[12],bool,address,bool))"))[:4]
	MethodIDGetBacking     = crypto.Keccak256([]byte("getBacking(address,uint256)"))[:4]
	MethodIDBurnAndRecover = crypto.Keccak256([]byte("burnAndRecover(address,uint256)"))[:4]
	MethodIDGetFloorPrice  = crypto.Keccak256([]byte("getFloorPrice(address)"))[:4]
)

// TokenConfig represents the configuration for creating an asset-backed token
// Note: All tokens are backed by Smart coin only (native coin)
type TokenConfig struct {
	Name          string
	Symbol        string
	TotalSupply   *big.Int
	// BackingAsset is always Smart coin (native coin) - address(0) or native
	// This field is kept for future compatibility but will be enforced as Smart
	BackingAsset  common.Address // Must be address(0) for Smart coin
	InitialBacking *big.Int      // Amount of Smart coin to lock as backing
	Fees          [12]*big.Int
	OnlySB        bool
	Owner         common.Address
	EnableLGE     bool
}

// BackingInfo represents backing information for a token
type BackingInfo struct {
	BackingAsset    common.Address
	TotalBacking    *big.Int
	TotalSupply     *big.Int
	BurnedSupply    *big.Int
	FloorPrice      *big.Int
	BackingPerToken *big.Int
}

// StatefulPrecompile is an extended interface for precompiles that need StateDB access
type StatefulPrecompile interface {
	vm.PrecompiledContract
	SetStateDB(stateDB vm.StateDB)
}

// Precompile implements the asset backing precompile
type Precompile struct {
	stateDB vm.StateDB
}

// NewPrecompile creates a new asset backing precompile instance
func NewPrecompile(stateDB vm.StateDB) *Precompile {
	return &Precompile{
		stateDB: stateDB,
	}
}

// SetStateDB sets the state database for the precompile
func (p *Precompile) SetStateDB(stateDB vm.StateDB) {
	p.stateDB = stateDB
}

// Name returns the precompile name
func (p *Precompile) Name() string {
	return "SmartDeFi Asset Backing"
}

// RequiredGas calculates the gas required for the precompile operation
func (p *Precompile) RequiredGas(input []byte) uint64 {
	if len(input) < 4 {
		return 0
	}
	
	methodID := input[:4]
	
	switch {
	case common.BytesToHash(methodID) == common.BytesToHash(MethodIDCreateToken):
		// Base cost + data size cost
		return GasCreateToken + uint64(len(input)-4)*GasPerByte
	case common.BytesToHash(methodID) == common.BytesToHash(MethodIDGetBacking):
		return GasGetBacking
	case common.BytesToHash(methodID) == common.BytesToHash(MethodIDBurnAndRecover):
		return GasBurnAndRecover
	case common.BytesToHash(methodID) == common.BytesToHash(MethodIDGetFloorPrice):
		return GasGetBacking
	default:
		return 0
	}
}

// Run executes the precompile logic (implements PrecompiledContract interface)
func (p *Precompile) Run(input []byte) ([]byte, error) {
	if p.stateDB == nil {
		return nil, vm.ErrExecutionReverted
	}
	
	if len(input) < 4 {
		return nil, vm.ErrExecutionReverted
	}
	
	methodID := input[:4]
	
	// For now, we'll need caller and value from EVM context
	// This is a simplified version - full implementation needs EVM modification
	switch {
	case common.BytesToHash(methodID) == common.BytesToHash(MethodIDCreateToken):
		// TODO: Get caller from EVM context
		caller := common.Address{} // Will be set by EVM
		return p.createAssetBackedToken(input[4:], caller, big.NewInt(0), false)
	case common.BytesToHash(methodID) == common.BytesToHash(MethodIDGetBacking):
		return p.getBacking(input[4:], true)
	case common.BytesToHash(methodID) == common.BytesToHash(MethodIDBurnAndRecover):
		caller := common.Address{} // Will be set by EVM
		return p.burnAndRecover(input[4:], caller, false)
	case common.BytesToHash(methodID) == common.BytesToHash(MethodIDGetFloorPrice):
		return p.getFloorPrice(input[4:], true)
	default:
		return nil, vm.ErrExecutionReverted
	}
}

// createAssetBackedToken creates a new asset-backed token natively on the chain
func (p *Precompile) createAssetBackedToken(input []byte, caller common.Address, value *big.Int, readOnly bool) ([]byte, error) {
	if readOnly {
		return nil, vm.ErrExecutionReverted
	}
	
	// Decode TokenConfig from input
	config, err := DecodeCreateTokenInput(input)
	if err != nil {
		return nil, vm.ErrExecutionReverted
	}
	
	// Validate configuration
	if err := validateTokenConfig(config); err != nil {
		return nil, vm.ErrExecutionReverted
	}
	
	// Enforce Smart coin as only backing asset
	// BackingAsset must be address(0) for native Smart coin
	if config.BackingAsset != (common.Address{}) {
		return nil, vm.ErrExecutionReverted // Only Smart coin supported
	}
	
	// Create deterministic token address (CREATE2-like)
	// Using caller address + nonce + config hash for determinism
	nonce := p.stateDB.GetNonce(caller)
	configHash := crypto.Keccak256Hash(
		caller.Bytes(),
		common.BigToHash(big.NewInt(int64(nonce))).Bytes(),
		[]byte(config.Name),
		[]byte(config.Symbol),
		config.TotalSupply.Bytes(),
	)
	tokenAddress := common.BytesToAddress(configHash[:20])
	
	// Check if token already exists
	if p.stateDB.GetCodeSize(tokenAddress) > 0 {
		return nil, vm.ErrExecutionReverted // Token already exists
	}
	
	// Initialize backing pool with Smart coin (native coin)
	// BackingAsset is always address(0) for Smart coin
	smartCoinAddress := common.Address{} // Native Smart coin
	
	pool := &backingpool.BackingPool{
		TokenAddress:  tokenAddress,
		BackingAsset:  smartCoinAddress, // Always Smart coin
		TotalBacking:  new(big.Int).Set(config.InitialBacking),
		TotalSupply:   new(big.Int).Set(config.TotalSupply),
		BurnedSupply:  big.NewInt(0),
		BackingAssets: []common.Address{smartCoinAddress}, // Only Smart coin
		BackingAmounts: []*big.Int{new(big.Int).Set(config.InitialBacking)},
	}
	
	// Save backing pool state
	backingpool.SetBackingPool(p.stateDB, pool)
	
	// Lock initial Smart coin backing (transfer from caller to precompile)
	// Smart coin is the native coin, so we transfer native balance
	if config.InitialBacking.Cmp(big.NewInt(0)) > 0 {
		// Transfer Smart coin from caller to precompile address
		// This locks the Smart coin as backing for the token
		p.stateDB.AddBalance(PrecompileAddressBytes, config.InitialBacking)
		p.stateDB.SubBalance(caller, config.InitialBacking)
	}
	
	// Store fee structure in state (using storage slots)
	storeFeeStructure(p.stateDB, tokenAddress, config.Fees, config.OnlySB)
	
	// Return token address (ABI encoded)
	return tokenAddress.Bytes(), nil
}

// validateTokenConfig validates the token configuration
func validateTokenConfig(config TokenConfig) error {
	// Validate supply
	if config.TotalSupply.Cmp(big.NewInt(0)) <= 0 {
		return vm.ErrExecutionReverted
	}
	
	// Validate fees (max 50% total)
	totalBuyFees := big.NewInt(0)
	totalSellFees := big.NewInt(0)
	for i := 0; i < 6; i++ {
		totalBuyFees.Add(totalBuyFees, config.Fees[i])
		totalSellFees.Add(totalSellFees, config.Fees[i+6])
	}
	
	if totalBuyFees.Cmp(big.NewInt(500)) > 0 || totalSellFees.Cmp(big.NewInt(500)) > 0 {
		return vm.ErrExecutionReverted // Max 50% fees
	}
	
	// Validate initial backing
	if config.InitialBacking.Cmp(big.NewInt(0)) < 0 {
		return vm.ErrExecutionReverted
	}
	
	return nil
}

// storeFeeStructure stores the fee structure in state
func storeFeeStructure(stateDB vm.StateDB, tokenAddress common.Address, fees [12]*big.Int, onlySB bool) {
	// Store fees in storage slots (simplified - actual implementation would use proper slot calculation)
	slotBase := getFeeSlotBase(tokenAddress)
	
	for i, fee := range fees {
		stateDB.SetState(tokenAddress, 
			common.BigToHash(big.NewInt(slotBase+int64(i))), 
			common.BigToHash(fee))
	}
	
	// Store onlySB flag
	onlySBValue := big.NewInt(0)
	if onlySB {
		onlySBValue = big.NewInt(1)
	}
	stateDB.SetState(tokenAddress, 
		common.BigToHash(big.NewInt(slotBase+12)), 
		common.BigToHash(onlySBValue))
}

func getFeeSlotBase(tokenAddress common.Address) int64 {
	hash := crypto.Keccak256Hash(tokenAddress.Bytes(), []byte("SmartDeFi-Fees"))
	return new(big.Int).Mod(hash.Big(), big.NewInt(1e10)).Int64()
}

// getBacking returns the backing information for a given token and amount
func (p *Precompile) getBacking(input []byte, readOnly bool) ([]byte, error) {
	// Decode input
	token, amount, err := DecodeGetBackingInput(input)
	if err != nil {
		return nil, vm.ErrExecutionReverted
	}
	
	// Get backing pool state
	pool := backingpool.GetBackingPool(p.stateDB, token)
	if pool == nil {
		return nil, vm.ErrExecutionReverted
	}
	
	// Calculate backing for amount
	backingAmount := pool.CalculateBackingForAmount(amount)
	
	// Return backing amount (ABI encoded)
	return EncodeOutput("getBacking", backingAmount)
}

// burnAndRecover burns tokens and recovers the backing assets
func (p *Precompile) burnAndRecover(input []byte, caller common.Address, readOnly bool) ([]byte, error) {
	if readOnly {
		return nil, vm.ErrExecutionReverted
	}
	
	// Decode input
	token, amount, err := DecodeBurnAndRecoverInput(input)
	if err != nil {
		return nil, vm.ErrExecutionReverted
	}
	
	// Get backing pool state
	pool := backingpool.GetBackingPool(p.stateDB, token)
	if pool == nil {
		return nil, vm.ErrExecutionReverted
	}
	
	// Verify caller has tokens (simplified - actual implementation needs token balance check)
	// For native tokens, we'd check balance from state
	// This is a placeholder - full implementation needs token contract integration
	
	// Calculate recoverable backing
	recoveredAmount := pool.CalculateBackingForAmount(amount)
	
	// Burn tokens (update burned supply)
	pool.BurnTokens(amount)
	
	// Update backing pool state
	pool.TotalBacking.Sub(pool.TotalBacking, recoveredAmount)
	backingpool.SetBackingPool(p.stateDB, pool)
	
	// Transfer Smart coin backing to caller
	// Smart coin is native, so we transfer native balance
	// BackingAsset is always address(0) for Smart coin
	if recoveredAmount.Cmp(big.NewInt(0)) > 0 {
		// Transfer Smart coin from precompile to caller
		p.stateDB.SubBalance(PrecompileAddressBytes, recoveredAmount)
		p.stateDB.AddBalance(caller, recoveredAmount)
	}
	
	// Return recovered amount (ABI encoded)
	return EncodeOutput("burnAndRecover", recoveredAmount)
}

// getFloorPrice returns the floor price for a token
func (p *Precompile) getFloorPrice(input []byte, readOnly bool) ([]byte, error) {
	// Decode input
	token, err := DecodeGetFloorPriceInput(input)
	if err != nil {
		return nil, vm.ErrExecutionReverted
	}
	
	// Get backing pool state
	pool := backingpool.GetBackingPool(p.stateDB, token)
	if pool == nil {
		return nil, vm.ErrExecutionReverted
	}
	
	// Calculate floor price
	floorPrice := pool.CalculateFloorPrice()
	
	// Return floor price (ABI encoded)
	return EncodeOutput("getFloorPrice", floorPrice)
}

