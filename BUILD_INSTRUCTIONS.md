# Building SmartDeFi Geth

## Prerequisites

- Go 1.21 or later
- Git (already installed)

## Build Methods

### Method 1: Using Make (Linux/Mac/WSL)

```bash
make geth
```

### Method 2: Using Go Directly (Windows/Linux/Mac)

```bash
go build ./cmd/geth
```

### Method 3: Cross-Platform Build

```bash
go build -o build/bin/geth ./cmd/geth
```

## After Building

The binary will be at:
- `build/bin/geth` (if using make)
- `geth.exe` (if using go build directly on Windows)

## Testing

```bash
go test ./core/vm/precompiles/assetbacking/... -v
go test ./core/state/backingpool/... -v
```

## Running

```bash
./build/bin/geth --dev
# or
./geth --dev
```

---

**Note:** On Windows, you may need to install Go first: https://go.dev/dl/

