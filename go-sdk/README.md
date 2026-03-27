# Trust Wallet Core Go SDK

A production-ready Go SDK for Trust Wallet Core supporting 164+ cryptocurrencies.

## Installation

```bash
go get github.com/trustwallet/go-wallet-core
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/trustwallet/go-wallet-core/pkg/coin"
    "github.com/trustwallet/go-wallet-core/pkg/wallet"
)

func main() {
    // Create wallet from mnemonic
    w, err := wallet.NewWalletFromMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
    if err != nil {
        log.Fatal(err)
    }

    // Derive Bitcoin account
    btcAccount, err := w.Derive(coin.Bitcoin)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("BTC Address: %s\n", btcAccount.Address())

    // Derive Ethereum account
    ethAccount, err := w.Derive(coin.Ethereum)
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("ETH Address: %s\n", ethAccount.Address())
}
```

## Features

- **164+ Coins Supported**: Full coverage of all Trust Wallet Core coins
- **Idiomatic Go API**: Clean, Go-style interfaces with proper error handling
- **Build Tags**: Works with or without CGO for testing
- **Type Safe**: Generated code ensures compile-time safety
- **Well Documented**: Comprehensive Go docs

## Supported Coins

The SDK supports all 164 coins from the Trust Wallet registry including:

- **Bitcoin Family**: Bitcoin, Litecoin, Dogecoin, Bitcoin Cash, etc.
- **Ethereum Family**: Ethereum, Polygon, Arbitrum, Optimism, 70+ EVM chains
- **Cosmos Family**: Cosmos, Osmosis, Juno, 20+ chains
- **Native Chains**: Solana, Cardano, Polkadot, Ripple, etc.

## Testing

Run tests without CGO (uses stub implementations):

```bash
CGO_ENABLED=0 go test ./...
```

Run tests with CGO (requires wallet-core library):

```bash
go test ./...
```

## Architecture

The SDK uses build tags to provide two implementations:

- **Without CGO** (`!cgo`): Stub implementations for testing
- **With CGO** (`cgo`): Full implementation using Trust Wallet Core C++ library

## Development

### Generate Coin Types

```bash
cd cmd/codegen
go run main.go ../../registry.json ../pkg/coin/
```

### Project Structure

```
go-sdk/
├── cmd/codegen/          # Code generation from registry.json
├── pkg/
│   ├── coin/            # Coin type definitions and utilities
│   └── wallet/          # Wallet operations
└── examples/            # Usage examples
```

## License

MIT License - see LICENSE file for details.
