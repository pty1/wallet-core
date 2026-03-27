//go:build !cgo
// +build !cgo

package transaction

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustwallet/go-wallet-core/pkg/coin"
)

func TestEthereumTransactionBuilder_Validation(t *testing.T) {
	tests := []struct {
		name    string
		builder func() *EthereumTransactionBuilder
		wantErr string
	}{
		{
			name: "missing recipient",
			builder: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction()
			},
			wantErr: "recipient address is required",
		},
		{
			name: "missing gas limit",
			builder: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction().To("0x1234567890123456789012345678901234567890")
			},
			wantErr: "gas limit is required",
		},
		{
			name: "valid legacy transaction",
			builder: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction().
					To("0x1234567890123456789012345678901234567890").
					GasLimit(21000).
					GasPrice(big.NewInt(1000000000)).
					Nonce(1).
					ChainID(big.NewInt(1))
			},
			wantErr: "",
		},
		{
			name: "valid EIP-1559 transaction",
			builder: func() *EthereumTransactionBuilder {
				return NewEthereumTransaction().
					Type(EthereumTxTypeEIP1559).
					To("0x1234567890123456789012345678901234567890").
					GasLimit(21000).
					MaxFeePerGas(big.NewInt(2000000000)).
					MaxPriorityFeePerGas(big.NewInt(1000000000)).
					Nonce(1).
					ChainID(big.NewInt(1))
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.builder().Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestEthereumTransactionBuilder_Sign_Stub(t *testing.T) {
	builder := NewEthereumTransaction().
		To("0x1234567890123456789012345678901234567890").
		GasLimit(21000).
		GasPrice(big.NewInt(1000000000)).
		Nonce(1).
		ChainID(big.NewInt(1))

	_, err := builder.Sign([]byte("private_key"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires CGO")
}

func TestBitcoinTransactionBuilder_Validation(t *testing.T) {
	tests := []struct {
		name    string
		builder func() *BitcoinTransactionBuilder
		wantErr string
	}{
		{
			name: "missing recipient",
			builder: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction()
			},
			wantErr: "recipient address is required",
		},
		{
			name: "missing change address",
			builder: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx")
			},
			wantErr: "change address is required",
		},
		{
			name: "zero amount",
			builder: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().
					To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx").
					Change("1FQc5LdgGHMHEN9nwkjmz6tWkxhPpxBvBU").
					Amount(0)
			},
			wantErr: "amount must be positive",
		},
		{
			name: "valid transaction",
			builder: func() *BitcoinTransactionBuilder {
				return NewBitcoinTransaction().
					To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx").
					Change("1FQc5LdgGHMHEN9nwkjmz6tWkxhPpxBvBU").
					Amount(100000).
					FeeRate(10).
					PrivateKeys([][]byte{[]byte("key")}).
					AddUTXO(BitcoinUTXO{
						TxHash:  []byte("hash"),
						TxIndex: 0,
						Amount:  200000,
						Script: []byte("script"),
					})
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.builder().Validate()
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestBitcoinTransactionBuilder_Sign_Stub(t *testing.T) {
	builder := NewBitcoinTransaction().
		To("1Bp9U1ogV3A14FMvKbRJms7ctyso4Z4Tcx").
		Change("1FQc5LdgGHMHEN9nwkjmz6tWkxhPpxBvBU").
		Amount(100000).
		FeeRate(10).
		PrivateKeys([][]byte{[]byte("key")}).
		AddUTXO(BitcoinUTXO{
			TxHash:  []byte("hash"),
			TxIndex: 0,
			Amount:  200000,
			Script: []byte("script"),
		})

	_, err := builder.Sign()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires CGO")
}

func TestBitcoinTransactionBuilder_CoinType(t *testing.T) {
	builder := NewBitcoinTransaction().CoinType(coin.Litecoin)
	assert.Equal(t, coin.Litecoin, builder.coinType)
}

func TestEthereumTransactionBuilder_ChainID(t *testing.T) {
	chainID := big.NewInt(1)
	builder := NewEthereumTransaction().ChainID(chainID)
	assert.Equal(t, chainID, builder.chainID)
}
