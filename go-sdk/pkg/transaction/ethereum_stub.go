//go:build !cgo
// +build !cgo

package transaction

import (
	"encoding/hex"
	"errors"
	"math/big"
)

// EthereumTxType represents the type of Ethereum transaction.
type EthereumTxType int

const (
	EthereumTxTypeLegacy EthereumTxType = iota
	EthereumTxTypeEIP1559
)

// EthereumTransactionBuilder helps construct Ethereum transactions.
type EthereumTransactionBuilder struct {
	to                   string
	value                *big.Int
	gasPrice             *big.Int
	gasLimit             uint64
	nonce                uint64
	chainID              *big.Int
	data                 []byte
	txType               EthereumTxType
	maxFeePerGas         *big.Int
	maxPriorityFeePerGas *big.Int
}

func NewEthereumTransaction() *EthereumTransactionBuilder {
	return &EthereumTransactionBuilder{
		value:    big.NewInt(0),
		gasPrice: big.NewInt(0),
		chainID:  big.NewInt(1),
		data:     []byte{},
		txType:   EthereumTxTypeLegacy,
	}
}

func (b *EthereumTransactionBuilder) To(address string) *EthereumTransactionBuilder {
	b.to = address
	return b
}

func (b *EthereumTransactionBuilder) Value(value *big.Int) *EthereumTransactionBuilder {
	b.value = new(big.Int).Set(value)
	return b
}

func (b *EthereumTransactionBuilder) GasPrice(price *big.Int) *EthereumTransactionBuilder {
	b.gasPrice = new(big.Int).Set(price)
	return b
}

func (b *EthereumTransactionBuilder) GasLimit(limit uint64) *EthereumTransactionBuilder {
	b.gasLimit = limit
	return b
}

func (b *EthereumTransactionBuilder) Nonce(nonce uint64) *EthereumTransactionBuilder {
	b.nonce = nonce
	return b
}

func (b *EthereumTransactionBuilder) ChainID(chainID *big.Int) *EthereumTransactionBuilder {
	b.chainID = new(big.Int).Set(chainID)
	return b
}

func (b *EthereumTransactionBuilder) Data(data []byte) *EthereumTransactionBuilder {
	b.data = data
	return b
}

func (b *EthereumTransactionBuilder) Type(txType EthereumTxType) *EthereumTransactionBuilder {
	b.txType = txType
	return b
}

func (b *EthereumTransactionBuilder) MaxFeePerGas(fee *big.Int) *EthereumTransactionBuilder {
	b.maxFeePerGas = new(big.Int).Set(fee)
	return b
}

func (b *EthereumTransactionBuilder) MaxPriorityFeePerGas(fee *big.Int) *EthereumTransactionBuilder {
	b.maxPriorityFeePerGas = new(big.Int).Set(fee)
	return b
}

func (b *EthereumTransactionBuilder) Validate() error {
	if b.to == "" {
		return errors.New("recipient address is required")
	}
	if b.gasLimit == 0 {
		return errors.New("gas limit is required")
	}
	if b.chainID == nil || b.chainID.Sign() <= 0 {
		return errors.New("chain ID is required")
	}
	return nil
}

func (b *EthereumTransactionBuilder) Sign(privateKey []byte) ([]byte, error) {
	return nil, errors.New("ethereum transaction signing requires CGO")
}

type EthereumTxResult struct {
	Raw       []byte
	Hash      string
	Signature string
}

func SignEthereumTransaction(
	privateKey []byte,
	chainID *big.Int,
	nonce uint64,
	gasLimit uint64,
	to string,
	value *big.Int,
	gasPrice *big.Int,
) (*EthereumTxResult, error) {
	return nil, errors.New("ethereum transaction signing requires CGO")
}

func SignEthereumTransactionEIP1559(
	privateKey []byte,
	chainID *big.Int,
	nonce uint64,
	gasLimit uint64,
	to string,
	value *big.Int,
	maxFeePerGas *big.Int,
	maxPriorityFeePerGas *big.Int,
) (*EthereumTxResult, error) {
	return nil, errors.New("ethereum transaction signing requires CGO")
}
