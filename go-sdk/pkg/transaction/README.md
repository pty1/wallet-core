# Transaction Package

This package provides transaction signing capabilities for the Trust Wallet Core SDK.

## API Design

```go
// High-level transaction API
tx := transaction.NewEthereumTransaction().
    To(address).
    Value(amount).
    GasPrice(gasPrice).
    GasLimit(gasLimit).
    Nonce(nonce).
    ChainID(chainID)

signed, err := tx.Sign(privateKey)
```
