//go:build cgo
// +build cgo

package transaction

/*
#include <stdlib.h>
#include <string.h>
#include <TrustWalletCore/TWCoinType.h>
#include <TrustWalletCore/TWAnySigner.h>
#include <TrustWalletCore/TWData.h>
#include <TrustWalletCore/TWString.h>
*/
import "C"
import (
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"unsafe"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	ethproto "github.com/trustwallet/go-wallet-core/pkg/proto/ethereum"
	"google.golang.org/protobuf/proto"
)

type EthereumTxType int

const (
	EthereumTxTypeLegacy EthereumTxType = iota
	EthereumTxTypeEIP1559
)

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

	if b.txType == EthereumTxTypeLegacy {
		if b.gasPrice == nil || b.gasPrice.Sign() <= 0 {
			return errors.New("gas price is required for legacy transactions")
		}
	} else if b.txType == EthereumTxTypeEIP1559 {
		if b.maxFeePerGas == nil || b.maxFeePerGas.Sign() <= 0 {
			return errors.New("maxFeePerGas is required for EIP-1559 transactions")
		}
		if b.maxPriorityFeePerGas == nil {
			b.maxPriorityFeePerGas = big.NewInt(0)
		}
	}

	return nil
}

func (b *EthereumTransactionBuilder) Sign(privateKey []byte) ([]byte, error) {
	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	input, err := b.buildSigningInput(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to build signing input: %w", err)
	}

	inputBytes, err := proto.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal input: %w", err)
	}

	cInputData := C.TWDataCreateWithBytes(
		(*C.uint8_t)(&inputBytes[0]),
		C.size_t(len(inputBytes)),
	)
	defer C.TWDataDelete(cInputData)

	cOutputData := C.TWAnySignerSign(cInputData, C.enum_TWCoinType(coin.Ethereum))
	defer C.TWDataDelete(cOutputData)

	outputSize := C.TWDataSize(cOutputData)
	outputBytes := make([]byte, outputSize)
	if outputSize > 0 {
		C.memcpy(
			unsafe.Pointer(&outputBytes[0]),
			unsafe.Pointer(C.TWDataBytes(cOutputData)),
			C.size_t(outputSize),
		)
	}

	return outputBytes, nil
}

func (b *EthereumTransactionBuilder) buildSigningInput(privateKey []byte) (*ethproto.SigningInput, error) {
	var txMode ethproto.TransactionMode
	if b.txType == EthereumTxTypeLegacy {
		txMode = ethproto.TransactionMode_Legacy
	} else {
		txMode = ethproto.TransactionMode_Enveloped
	}

	input := &ethproto.SigningInput{
		ChainId:    b.chainID.Bytes(),
		Nonce:      big.NewInt(int64(b.nonce)).Bytes(),
		GasLimit:   big.NewInt(int64(b.gasLimit)).Bytes(),
		ToAddress:  b.to,
		PrivateKey: privateKey,
		TxMode:     txMode,
	}

	if b.txType == EthereumTxTypeLegacy {
		input.GasPrice = b.gasPrice.Bytes()
	} else {
		input.MaxFeePerGas = b.maxFeePerGas.Bytes()
		input.MaxInclusionFeePerGas = b.maxPriorityFeePerGas.Bytes()
	}

	input.Transaction = &ethproto.Transaction{
		TransactionOneof: &ethproto.Transaction_Transfer_{
			Transfer: &ethproto.Transaction_Transfer{
				Amount: b.value.Bytes(),
				Data:   b.data,
			},
		},
	}

	return input, nil
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
	tx := NewEthereumTransaction().
		ChainID(chainID).
		Nonce(nonce).
		GasLimit(gasLimit).
		To(to).
		Value(value).
		GasPrice(gasPrice)

	raw, err := tx.Sign(privateKey)
	if err != nil {
		return nil, err
	}

	return &EthereumTxResult{
		Raw:       raw,
		Hash:      hex.EncodeToString(raw),
		Signature: "",
	}, nil
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
	tx := NewEthereumTransaction().
		Type(EthereumTxTypeEIP1559).
		ChainID(chainID).
		Nonce(nonce).
		GasLimit(gasLimit).
		To(to).
		Value(value).
		MaxFeePerGas(maxFeePerGas).
		MaxPriorityFeePerGas(maxPriorityFeePerGas)

	raw, err := tx.Sign(privateKey)
	if err != nil {
		return nil, err
	}

	return &EthereumTxResult{
		Raw:  raw,
		Hash: hex.EncodeToString(raw),
	}, nil
}
