# SmartDeFi L2 - Geth Fork

This is a fork of go-ethereum with the SmartDeFi asset backing precompile integrated.

## What's Added

- **SmartDeFi Precompile** at address `0x0000000000000000000000000000000000000100`
- **Smart Coin Enforcement** - All tokens backed by Smart coin only
- **Protocol-Level Backing** - Backing pools managed at chain level
- **Native Token Creation** - Create asset-backed tokens natively

## Building

```bash
make geth
```

## Testing

```bash
go test ./core/vm/precompiles/assetbacking/... -v
go test ./core/state/backingpool/... -v
```

## Usage with OP Stack

This forked geth is designed to be used as the execution client for OP Stack L2.

## Precompile Address

`0x0000000000000000000000000000000000000100`

## Features

- ✅ Native asset-backed token creation
- ✅ Smart coin as only backing asset
- ✅ Guaranteed floor price
- ✅ Burn-to-recover mechanism
- ✅ Protocol-level state management

---

**Original:** https://github.com/ethereum/go-ethereum  
**Fork:** https://github.com/Dtheapp/go-ethereum

