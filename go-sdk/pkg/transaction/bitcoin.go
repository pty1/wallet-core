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
#include <TrustWalletCore/TWPublicKey.h>
#include <TrustWalletCore/TWPrivateKey.h>
*/
import "C"
import (
	"encoding/hex"
	"errors"
	"fmt"
	"unsafe"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	btcproto "github.com/trustwallet/go-wallet-core/pkg/proto/bitcoin"
	"github.com/trustwallet/go-wallet-core/pkg/proto/common"
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

func (u *BitcoinUTXO) toProto() *btcproto.UnspentTransaction {
	return &btcproto.UnspentTransaction{
		OutPoint: &btcproto.OutPoint{
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
	result, err := b.SignWithResult()
	if err != nil {
		return nil, err
	}
	return result.Encoded, nil
}

func (b *BitcoinTransactionBuilder) SignWithResult() (*BitcoinTxResult, error) {
	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	protoUTXOs := make([]*btcproto.UnspentTransaction, len(b.utxos))
	for i, utxo := range b.utxos {
		protoUTXOs[i] = utxo.toProto()
	}

	hashType := uint32(C.TWBitcoinScriptHashTypeForCoin(C.enum_TWCoinType(b.coinType)))

	input := &btcproto.SigningInput{
		HashType:      hashType,
		Amount:        b.amount,
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
	if outputSize == 0 {
		return nil, errors.New("signing returned empty result")
	}

	outputBytes := make([]byte, outputSize)
	C.memcpy(
		unsafe.Pointer(&outputBytes[0]),
		unsafe.Pointer(C.TWDataBytes(cOutputData)),
		C.size_t(outputSize),
	)

	var output btcproto.SigningOutput
	if err := proto.Unmarshal(outputBytes, &output); err != nil {
		return nil, fmt.Errorf("failed to unmarshal signing output: %w", err)
	}

	if output.Error != common.SigningError_OK {
		errMsg := output.ErrorMessage
		if errMsg == "" {
			errMsg = fmt.Sprintf("signing error code: %d", output.Error)
		}
		return nil, fmt.Errorf("transaction signing failed: %s", errMsg)
	}

	encoded := output.GetEncoded()
	if len(encoded) == 0 {
		return nil, errors.New("signed transaction is empty")
	}

	return &BitcoinTxResult{
		Encoded:       encoded,
		TransactionID: output.GetTransactionId(),
		Transaction:   output.GetTransaction(),
	}, nil
}

type BitcoinTxResult struct {
	Encoded       []byte
	TransactionID string
	Transaction   *btcproto.Transaction
}

func BuildLockScript(address string, coinType coin.CoinType) ([]byte, error) {
	cAddress := C.TWStringCreateWithUTF8Bytes(C.CString(address))
	defer C.TWStringDelete(cAddress)

	script := C.TWBitcoinScriptLockScriptForAddress(cAddress, C.enum_TWCoinType(coinType))
	defer C.TWBitcoinScriptDelete(script)

	scriptData := C.TWBitcoinScriptData(script)
	defer C.TWDataDelete(scriptData)

	size := C.TWDataSize(scriptData)
	if size == 0 {
		return nil, errors.New("failed to build lock script")
	}

	bytes := make([]byte, size)
	C.memcpy(
		unsafe.Pointer(&bytes[0]),
		unsafe.Pointer(C.TWDataBytes(scriptData)),
		C.size_t(size),
	)

	return bytes, nil
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

func PrivateKeyToPublicKey(privateKey []byte) ([]byte, error) {
	if len(privateKey) != 32 {
		return nil, errors.New("invalid private key length")
	}

	cPrivKey := C.TWDataCreateWithBytes(
		(*C.uint8_t)(&privateKey[0]),
		C.size_t(len(privateKey)),
	)
	defer C.TWDataDelete(cPrivKey)

	twPrivKey := C.TWPrivateKeyCreateWithData(cPrivKey)
	defer C.TWPrivateKeyDelete(twPrivKey)

	twPubKey := C.TWPrivateKeyGetPublicKeySecp256k1(twPrivKey, true)
	defer C.TWPublicKeyDelete(twPubKey)

	pubKeyData := C.TWPublicKeyData(twPubKey)
	defer C.TWDataDelete(pubKeyData)

	size := C.TWDataSize(pubKeyData)
	if size == 0 {
		return nil, errors.New("failed to get public key")
	}

	bytes := make([]byte, size)
	C.memcpy(
		unsafe.Pointer(&bytes[0]),
		unsafe.Pointer(C.TWDataBytes(pubKeyData)),
		C.size_t(size),
	)

	return bytes, nil
}

func ParseHexPrivateKey(hexKey string) ([]byte, error) {
	return hex.DecodeString(hexKey)
}
