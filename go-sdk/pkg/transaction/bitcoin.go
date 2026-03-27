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
#include <TrustWalletCore/TWBitcoinScript.h>
*/
import "C"
import (
	"encoding/hex"
	"errors"
	"fmt"
	"unsafe"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	bitcoinproto "github.com/trustwallet/go-wallet-core/pkg/proto/bitcoin"
	"google.golang.org/protobuf/proto"
)

type BitcoinSigHashType int

const (
	BitcoinSigHashTypeAll BitcoinSigHashType = iota + 1
	BitcoinSigHashTypeNone
	BitcoinSigHashTypeSingle
	BitcoinSigHashTypeAllAnyoneCanPay
	BitcoinSigHashTypeNoneAnyoneCanPay
	BitcoinSigHashTypeSingleAnyoneCanPay
)

type BitcoinUTXO struct {
	TxHash   []byte
	TxIndex  uint32
	Amount   int64
	Script   []byte
	Sequence uint32
}

func (u *BitcoinUTXO) toProto() *bitcoinproto.UnspentTransaction {
	return &bitcoinproto.UnspentTransaction{
		OutPoint: &bitcoinproto.OutPoint{
			Hash:     u.TxHash,
			Index:    u.TxIndex,
			Sequence: u.Sequence,
		},
		Amount: u.Amount,
		Script: u.Script,
	}
}

type BitcoinTransactionBuilder struct {
	coinType      coin.CoinType
	toAddress     string
	changeAddress string
	amount        int64
	feeRate       int64
	privateKeys   [][]byte
	utxos         []BitcoinUTXO
	sigHashType   BitcoinSigHashType
}

func NewBitcoinTransaction() *BitcoinTransactionBuilder {
	return &BitcoinTransactionBuilder{
		coinType:    coin.Bitcoin,
		sigHashType: BitcoinSigHashTypeAll,
	}
}

func (b *BitcoinTransactionBuilder) CoinType(ct coin.CoinType) *BitcoinTransactionBuilder {
	b.coinType = ct
	return b
}

func (b *BitcoinTransactionBuilder) To(address string) *BitcoinTransactionBuilder {
	b.toAddress = address
	return b
}

func (b *BitcoinTransactionBuilder) Change(address string) *BitcoinTransactionBuilder {
	b.changeAddress = address
	return b
}

func (b *BitcoinTransactionBuilder) Amount(amount int64) *BitcoinTransactionBuilder {
	b.amount = amount
	return b
}

func (b *BitcoinTransactionBuilder) FeeRate(rate int64) *BitcoinTransactionBuilder {
	b.feeRate = rate
	return b
}

func (b *BitcoinTransactionBuilder) PrivateKeys(keys [][]byte) *BitcoinTransactionBuilder {
	b.privateKeys = make([][]byte, len(keys))
	copy(b.privateKeys, keys)
	return b
}

func (b *BitcoinTransactionBuilder) AddUTXO(utxo BitcoinUTXO) *BitcoinTransactionBuilder {
	b.utxos = append(b.utxos, utxo)
	return b
}

func (b *BitcoinTransactionBuilder) SigHashType(ht BitcoinSigHashType) *BitcoinTransactionBuilder {
	b.sigHashType = ht
	return b
}

func (b *BitcoinTransactionBuilder) Validate() error {
	if b.toAddress == "" {
		return errors.New("recipient address is required")
	}
	if b.changeAddress == "" {
		return errors.New("change address is required")
	}
	if b.amount <= 0 {
		return errors.New("amount must be positive")
	}
	if b.feeRate <= 0 {
		return errors.New("fee rate must be positive")
	}
	if len(b.utxos) == 0 {
		return errors.New("at least one UTXO is required")
	}
	if len(b.privateKeys) == 0 {
		return errors.New("at least one private key is required")
	}
	return nil
}

func (b *BitcoinTransactionBuilder) Sign() ([]byte, error) {
	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	protoUTXOs := make([]*bitcoinproto.UnspentTransaction, len(b.utxos))
	for i, utxo := range b.utxos {
		protoUTXOs[i] = utxo.toProto()
	}

	input := &bitcoinproto.SigningInput{
		HashType:      uint32(b.sigHashType),
		Amount:       b.amount,
		ByteFee:       b.feeRate,
		ToAddress:     b.toAddress,
		ChangeAddress: b.changeAddress,
		PrivateKey:    b.privateKeys,
		Utxo:          protoUTXOs,
		CoinType:      uint32(b.coinType),
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

	cOutputData := C.TWAnySignerSign(cInputData, C.enum_TWCoinType(b.coinType))
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

func (b *BitcoinTransactionBuilder) SignWithResult() (*BitcoinTxResult, error) {
	raw, err := b.Sign()
	if err != nil {
		return nil, err
	}

	return &BitcoinTxResult{
		Raw:  raw,
		Hash: hex.EncodeToString(raw),
	}, nil
}

type BitcoinTxResult struct {
	Raw  []byte
	Hash string
}

func SignBitcoinTransaction(
	coinType coin.CoinType,
	privateKeys [][]byte,
	utxos []BitcoinUTXO,
	toAddress string,
	changeAddress string,
	amount int64,
	feeRate int64,
) (*BitcoinTxResult, error) {
	builder := NewBitcoinTransaction().
		CoinType(coinType).
		To(toAddress).
		Change(changeAddress).
		Amount(amount).
		FeeRate(feeRate).
		PrivateKeys(privateKeys)

	for _, utxo := range utxos {
		builder.AddUTXO(utxo)
	}

	return builder.SignWithResult()
}
