//go:build !cgo
// +build !cgo

// Package wallet provides wallet operations for the Trust Wallet Core SDK.
// This is a stub implementation for testing without CGO.
package wallet

import (
	"errors"
	"fmt"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
)

// Wallet represents a hierarchical deterministic wallet
type Wallet struct {
	mnemonic string
}

// NewWalletFromMnemonic creates a new wallet from a mnemonic phrase.
func NewWalletFromMnemonic(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic cannot be empty")
	}
	// In real implementation, validate mnemonic
	return &Wallet{mnemonic: mnemonic}, nil
}

// Derive derives an account for the specified coin type.
func (w *Wallet) Derive(ct coin.CoinType) (*Account, error) {
	if w.mnemonic == "" {
		return nil, errors.New("wallet not initialized")
	}
	// In real implementation, derive keys from mnemonic
	return &Account{
		coinType: ct,
		address:  fmt.Sprintf("stub_%s_address", ct.Symbol()),
		pubKey:   "stub_public_key",
		priKey:   "stub_private_key",
	}, nil
}

// Account represents a derived account for a specific coin
type Account struct {
	coinType coin.CoinType
	address  string
	pubKey   string
	priKey   string
}

// Address returns the account address
func (a *Account) Address() string {
	return a.address
}

// PublicKey returns the account public key
func (a *Account) PublicKey() string {
	return a.pubKey
}

// CoinType returns the coin type of the account
func (a *Account) CoinType() coin.CoinType {
	return a.coinType
}

// SignTransaction signs a transaction
func (a *Account) SignTransaction(data []byte) ([]byte, error) {
	return nil, errors.New("SignTransaction not implemented in stub")
}
