// Package backingpool manages protocol-level backing pool state
package backingpool

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
)

const (
	// Storage slot prefixes for backing pool state
	// Slot layout:
	// slot[0] = totalBacking
	// slot[1] = totalSupply
	// slot[2] = burnedSupply
	// slot[3] = backingAsset address
	// slot[4+] = additional backing assets (multi-asset support)
	
	SlotTotalBacking  = 0
	SlotTotalSupply   = 1
	SlotBurnedSupply  = 2
	SlotBackingAsset  = 3
	SlotBackingAssets = 4 // Array start
)

// BackingPool represents the protocol-level backing pool for a token
type BackingPool struct {
	TokenAddress    common.Address
	BackingAsset    common.Address
	TotalBacking    *big.Int
	TotalSupply     *big.Int
	BurnedSupply    *big.Int
	BackingAssets   []common.Address  // Multi-asset backing support
	BackingAmounts  []*big.Int
}

// GetBackingPool retrieves backing pool state from the state database
func GetBackingPool(stateDB *state.StateDB, tokenAddress common.Address) *BackingPool {
	// Calculate storage slots for this token
	// Using CREATE2-like deterministic slot calculation
	slotBase := getSlotBase(tokenAddress)
	
	// Read state from slots
	totalBacking := stateDB.GetState(tokenAddress, common.BigToHash(big.NewInt(slotBase+SlotTotalBacking))).Big()
	totalSupply := stateDB.GetState(tokenAddress, common.BigToHash(big.NewInt(slotBase+SlotTotalSupply))).Big()
	burnedSupply := stateDB.GetState(tokenAddress, common.BigToHash(big.NewInt(slotBase+SlotBurnedSupply))).Big()
	backingAssetBytes := stateDB.GetState(tokenAddress, common.BigToHash(big.NewInt(slotBase+SlotBackingAsset))).Bytes()
	backingAsset := common.BytesToAddress(backingAssetBytes[12:])
	
	// TODO: Read multi-asset backing arrays
	
	return &BackingPool{
		TokenAddress: tokenAddress,
		BackingAsset: backingAsset,
		TotalBacking: totalBacking,
		TotalSupply:  totalSupply,
		BurnedSupply: burnedSupply,
	}
}

// SetBackingPool writes backing pool state to the state database
func SetBackingPool(stateDB *state.StateDB, pool *BackingPool) {
	slotBase := getSlotBase(pool.TokenAddress)
	
	// Write state to slots
	stateDB.SetState(pool.TokenAddress, 
		common.BigToHash(big.NewInt(slotBase+SlotTotalBacking)), 
		common.BigToHash(pool.TotalBacking))
	
	stateDB.SetState(pool.TokenAddress, 
		common.BigToHash(big.NewInt(slotBase+SlotTotalSupply)), 
		common.BigToHash(pool.TotalSupply))
	
	stateDB.SetState(pool.TokenAddress, 
		common.BigToHash(big.NewInt(slotBase+SlotBurnedSupply)), 
		common.BigToHash(pool.BurnedSupply))
	
	// Write backing asset address (padded to 32 bytes)
	backingAssetHash := common.BigToHash(new(big.Int).SetBytes(pool.BackingAsset.Bytes()))
	stateDB.SetState(pool.TokenAddress, 
		common.BigToHash(big.NewInt(slotBase+SlotBackingAsset)), 
		backingAssetHash)
	
	// TODO: Write multi-asset backing arrays
}

// CalculateFloorPrice calculates the floor price per token
func (p *BackingPool) CalculateFloorPrice() *big.Int {
	if p.TotalSupply.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	
	// Floor price = Total Backing / (Total Supply - Burned Supply)
	circulatingSupply := new(big.Int).Sub(p.TotalSupply, p.BurnedSupply)
	if circulatingSupply.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	
	// Multiply by 1e18 for precision, then divide
	floorPrice := new(big.Int).Mul(p.TotalBacking, big.NewInt(1e18))
	floorPrice.Div(floorPrice, circulatingSupply)
	
	return floorPrice
}

// CalculateBackingForAmount calculates how much backing is available for a given token amount
func (p *BackingPool) CalculateBackingForAmount(amount *big.Int) *big.Int {
	if p.TotalSupply.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	
	circulatingSupply := new(big.Int).Sub(p.TotalSupply, p.BurnedSupply)
	if circulatingSupply.Cmp(big.NewInt(0)) == 0 {
		return big.NewInt(0)
	}
	
	// backing = (amount * totalBacking) / circulatingSupply
	backing := new(big.Int).Mul(amount, p.TotalBacking)
	backing.Div(backing, circulatingSupply)
	
	return backing
}

// AddBacking adds backing to the pool (from transaction fees)
func (p *BackingPool) AddBacking(amount *big.Int) {
	p.TotalBacking.Add(p.TotalBacking, amount)
}

// BurnTokens burns tokens and updates the pool state
func (p *BackingPool) BurnTokens(amount *big.Int) {
	p.BurnedSupply.Add(p.BurnedSupply, amount)
}

// getSlotBase calculates the base storage slot for a token's backing pool
func getSlotBase(tokenAddress common.Address) int64 {
	// Use token address to deterministically calculate slot base
	// This ensures each token has unique storage slots
	hash := common.Keccak256Hash(tokenAddress.Bytes(), []byte("SmartDeFi-BackingPool"))
	return new(big.Int).Mod(hash.Big(), big.NewInt(1e10)).Int64()
}

