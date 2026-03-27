// Package transaction provides transaction signing capabilities for the Trust Wallet Core SDK.
// It supports multiple blockchain types including Ethereum, Bitcoin, and others.
package transaction

import (
	"context"
	"errors"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
)

// Transaction represents a blockchain transaction that can be signed.
type Transaction interface {
	CoinType() coin.CoinType
	Serialize() ([]byte, error)
	SetSignature(signature []byte) error
}

// Signer is responsible for signing transactions.
type Signer interface {
	SignTransaction(ctx context.Context, tx Transaction) ([]byte, error)
}

// PrivateKeySigner implements Signer using a private key.
type PrivateKeySigner struct {
	privateKey []byte
}

// NewPrivateKeySigner creates a new signer from a private key.
func NewPrivateKeySigner(privateKey []byte) *PrivateKeySigner {
	return &PrivateKeySigner{privateKey: privateKey}
}

// SignTransaction signs a transaction using the private key.
func (s *PrivateKeySigner) SignTransaction(ctx context.Context, tx Transaction) ([]byte, error) {
	if len(s.privateKey) == 0 {
		return nil, errors.New("private key is empty")
	}
	return nil, errors.New("SignTransaction not yet implemented")
}

var (
	ErrInvalidTransaction = errors.New("invalid transaction")
	ErrSigningFailed      = errors.New("transaction signing failed")
	ErrUnsupportedCoin    = errors.New("unsupported coin type")
)
