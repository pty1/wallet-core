//go:build cgo
// +build cgo

package transaction

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustwallet/go-wallet-core/pkg/coin"
)

func TestEthereumTransaction_Sign_Legacy(t *testing.T) {
	privateKeyHex := "4c4ab1e51c1f05e3b0c6b5d0e3f7a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8"
	privateKey, err := hex.DecodeString(privateKeyHex)
	require.NoError(t, err)

	chainID := big.NewInt(1)
	value := big.NewInt(1000000000000000000) // 1 ETH
	gasPrice := big.NewInt(1000000000)       // 1 Gwei

	tx := NewEthereumTransaction().
		ChainID(chainID).
		Nonce(0).
		GasLimit(21000).
		To("0x1234567890123456789012345678901234567890").
		Value(value).
		GasPrice(gasPrice)

	signed, err := tx.Sign(privateKey)
	require.NoError(t, err)
	assert.NotEmpty(t, signed)
	assert.Greater(t, len(signed), 50) // Should have substantial data
}

func TestEthereumTransaction_Sign_EIP1559(t *testing.T) {
	privateKeyHex := "4c4ab1e51c1f05e3b0c6b5d0e3f7a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8"
	privateKey, err := hex.DecodeString(privateKeyHex)
	require.NoError(t, err)

	chainID := big.NewInt(1)
	value := big.NewInt(1000000000000000000) // 1 ETH
	maxFeePerGas := big.NewInt(2000000000)   // 2 Gwei
	maxPriorityFeePerGas := big.NewInt(1000000000) // 1 Gwei

	tx := NewEthereumTransaction().
		Type(EthereumTxTypeEIP1559).
		ChainID(chainID).
		Nonce(0).
		GasLimit(21000).
		To("0x1234567890123456789012345678901234567890").
		Value(value).
		MaxFeePerGas(maxFeePerGas).
		MaxPriorityFeePerGas(maxPriorityFeePerGas)

	signed, err := tx.Sign(privateKey)
	require.NoError(t, err)
	assert.NotEmpty(t, signed)
}

func TestEthereumTransaction_Validation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *EthereumTransactionBuilder
		wantErr string
	}{
		{
			name: "missing recipient",
			setup: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction().
					ChainID(big.NewInt(1)).
					GasLimit(21000).
					GasPrice(big.NewInt(1000000000))
			},
			wantErr: "recipient address is required",
		},
		{
			name: "missing gas limit",
			setup: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction().
					ChainID(big.NewInt(1)).
					To("0x1234567890123456789012345678901234567890").
					GasPrice(big.NewInt(1000000000))
			},
			wantErr: "gas limit is required",
		},
		{
			name: "missing chain ID",
			setup: func() *EthereumTransactionBuilder {
				tx := NewEthereumTransaction().
					To("0x1234567890123456789012345678901234567890").
					GasLimit(21000).
					GasPrice(big.NewInt(1000000000))
				tx.chainID = nil
				return tx
			},
			wantErr: "chain ID is required",
		},
		{
			name: "legacy missing gas price",
			setup: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction().
					Type(EthereumTxTypeLegacy).
					ChainID(big.NewInt(1)).
					To("0x1234567890123456789012345678901234567890").
					GasLimit(21000)
			},
			wantErr: "gas price is required for legacy transactions",
		},
		{
			name: "EIP1559 missing maxFeePerGas",
			setup: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction().
					Type(EthereumTxTypeEIP1559).
					ChainID(big.NewInt(1)).
					To("0x1234567890123456789012345678901234567890").
					GasLimit(21000)
			},
			wantErr: "maxFeePerGas is required for EIP-1559 transactions",
		},
		{
			name: "valid legacy transaction",
			setup: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction().
					ChainID(big.NewInt(1)).
					To("0x1234567890123456789012345678901234567890").
					GasLimit(21000).
					GasPrice(big.NewInt(1000000000)).
					Nonce(0).
					Value(big.NewInt(1000000000000000000))
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			err := builder.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBitcoinTransaction_Validation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func() *BitcoinTransactionBuilder
		wantErr string
	}{
		{
			name: "missing recipient",
			setup: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction()
			},
			wantErr: "recipient address is required",
		},
		{
			name: "missing change address",
			setup: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx")
			},
			wantErr: "change address is required",
		},
		{
			name: "zero amount",
			setup: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().
					To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx").
					Change("1FQc5LdgGHMHEN9nwkjmz6tWkxhPpxBvBU").
					Amount(0)
			},
			wantErr: "amount must be positive",
		},
		{
			name: "zero fee rate",
			setup: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().
					To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx").
					Change("1FQc5LdgGHMHEN9nwkjmz6tWkxhPpxBvBU").
					Amount(100000).
					FeeRate(0)
			},
			wantErr: "fee rate must be positive",
		},
		{
			name: "no UTXOs",
			setup: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().
					To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx").
					Change("1FQc5LdgGHMHEN9nwkjmz6tWkxhPpxBvBU").
					Amount(100000).
					FeeRate(10).
					PrivateKeys([][]byte{[]byte("key")})
			},
			wantErr: "at least one UTXO is required",
		},
		{
			name: "no private keys",
			setup: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().
					To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx").
					Change("1FQc5LdgGHMHEN9nwkjmz6tWkxhPpxBvBU").
					Amount(100000).
					FeeRate(10).
					AddUTXO(BitcoinUTXO{
						TxHash:  []byte("hash"),
						Amount:  200000,
						Script:  []byte("script"),
						TxIndex: 0,
					})
			},
			wantErr: "at least one private key is required",
		},
		{
			name: "valid setup",
			setup: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().
					To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx").
					Change("1FQc5LdgGHMHEN9nwkjmz6tWkxhPpxBvBU").
					Amount(100000).
					FeeRate(10).
					PrivateKeys([][]byte{[]byte("key")}).
					AddUTXO(BitcoinUTXO{
						TxHash:  []byte("hash"),
						Amount:  200000,
						Script:  []byte("script"),
						TxIndex: 0,
					})
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := tt.setup()
			err := builder.Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBitcoinTransaction_CoinType(t *testing.T) {
	tests := []struct {
		name    string
		coin    coin.CoinType
	}{
		{"Bitcoin", coin.Bitcoin},
		{"Litecoin", coin.Litecoin},
		{"Dogecoin", coin.Doge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			builder := NewBitcoinTransaction().CoinType(tt.coin)
			assert.Equal(t, tt.coin, builder.coinType)
		})
	}
}
