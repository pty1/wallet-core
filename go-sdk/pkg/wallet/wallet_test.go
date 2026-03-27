package wallet

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustwallet/go-wallet-core/pkg/coin"
)

func TestNewWalletFromMnemonic(t *testing.T) {
	tests := []struct {
		name      string
		mnemonic  string
		wantError bool
	}{
		{
			name:      "valid mnemonic",
			mnemonic:  "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about",
			wantError: false,
		},
		{
			name:      "empty mnemonic",
			mnemonic:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wallet, err := NewWalletFromMnemonic(tt.mnemonic)
			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, wallet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, wallet)
			}
		})
	}
}

func TestWallet_Derive(t *testing.T) {
	wallet, err := NewWalletFromMnemonic("abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about")
	require.NoError(t, err)

	tests := []struct {
		name     string
		coinType coin.CoinType
	}{
		{"derive bitcoin", coin.Bitcoin},
		{"derive ethereum", coin.Ethereum},
		{"derive litecoin", coin.Litecoin},
		{"derive doge", coin.Doge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := wallet.Derive(tt.coinType)
			require.NoError(t, err)
			assert.NotNil(t, account)
			assert.NotEmpty(t, account.Address())
			assert.Equal(t, tt.coinType, account.CoinType())
			assert.NotEmpty(t, account.PublicKey())
			assert.NotEmpty(t, account.PrivateKey())
		})
	}
}

func TestAccount_Address(t *testing.T) {
	acc := &Account{
		address: "test_address_123",
	}
	assert.Equal(t, "test_address_123", acc.Address())
}

func TestAccount_PublicKey(t *testing.T) {
	acc := &Account{
		pubKey: "test_pubkey_456",
	}
	assert.Equal(t, "test_pubkey_456", acc.PublicKey())
}

func TestAccount_CoinType(t *testing.T) {
	acc := &Account{
		coinType: coin.Ethereum,
	}
	assert.Equal(t, coin.Ethereum, acc.CoinType())
}

func TestAccount_SignTransaction(t *testing.T) {
	acc := &Account{}
	_, err := acc.SignTransaction([]byte("test"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "transaction package")
}
