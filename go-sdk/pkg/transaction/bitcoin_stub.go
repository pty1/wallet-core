//go:build !cgo
// +build !cgo

package transaction

import (
	"encoding/hex"
	"errors"
	"math/big"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
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
	return nil
}

func (b *BitcoinTransactionBuilder) Sign() ([]byte, error) {
	return nil, errors.New("bitcoin transaction signing requires CGO")
}

func (b *BitcoinTransactionBuilder) SignWithResult() (*BitcoinTxResult, error) {
	return nil, errors.New("bitcoin transaction signing requires CGO")
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
	return nil, errors.New("bitcoin transaction signing requires CGO")
}
