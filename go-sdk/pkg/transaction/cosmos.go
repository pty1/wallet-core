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
#include <TrustWalletCore/TWPublicKey.h>
#include <TrustWalletCore/TWPrivateKey.h>
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	"github.com/trustwallet/go-wallet-core/pkg/proto/common"
	cosmosproto "github.com/trustwallet/go-wallet-core/pkg/proto/cosmos"
	"google.golang.org/protobuf/proto"
)

type CosmosChainConfig struct {
	ChainID       string
	Denom         string
	Hrp           string
	CoinType      coin.CoinType
	PublicKeyType cosmosproto.SignerPublicKeyType
}

var CosmosChainConfigs = map[coin.CoinType]CosmosChainConfig{
	coin.Cosmos:          {ChainID: "cosmoshub-4", Denom: "uatom", Hrp: "cosmos", CoinType: coin.Cosmos, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Stargaze:        {ChainID: "stargaze-1", Denom: "ustars", Hrp: "stars", CoinType: coin.Stargaze, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Juno:            {ChainID: "juno-1", Denom: "ujuno", Hrp: "juno", CoinType: coin.Juno, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Stride:          {ChainID: "stride-1", Denom: "ustrd", Hrp: "stride", CoinType: coin.Stride, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Axelar:          {ChainID: "axelar-dojo-1", Denom: "uaxl", Hrp: "axelar", CoinType: coin.Axelar, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Crescent:        {ChainID: "crescent-1", Denom: "ucre", Hrp: "cre", CoinType: coin.Crescent, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Kujira:          {ChainID: "kaiyo-1", Denom: "ukuji", Hrp: "kujira", CoinType: coin.Kujira, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Comdex:          {ChainID: "comdex-1", Denom: "ucmdx", Hrp: "comdex", CoinType: coin.Comdex, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Neutron:         {ChainID: "neutron-1", Denom: "untrn", Hrp: "neutron", CoinType: coin.Neutron, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Sommelier:       {ChainID: "sommelier-3", Denom: "usomm", Hrp: "somm", CoinType: coin.Sommelier, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Fetchai:         {ChainID: "fetchhub-1", Denom: "afet", Hrp: "fetch", CoinType: coin.Fetchai, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Mars:            {ChainID: "mars-1", Denom: "umars", Hrp: "mars", CoinType: coin.Mars, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Umee:            {ChainID: "umee-1", Denom: "uumee", Hrp: "umee", CoinType: coin.Umee, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Noble:           {ChainID: "noble-1", Denom: "uusdc", Hrp: "noble", CoinType: coin.Noble, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Sei:             {ChainID: "sei-1", Denom: "usei", Hrp: "sei", CoinType: coin.Sei, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Tia:             {ChainID: "celestia", Denom: "utia", Hrp: "celestia", CoinType: coin.Tia, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Coreum:          {ChainID: "coreum-mainnet-1", Denom: "ucore", Hrp: "core", CoinType: coin.Coreum, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Quasar:          {ChainID: "quasar-1", Denom: "uqsr", Hrp: "quasar", CoinType: coin.Quasar, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Persistence:     {ChainID: "core-1", Denom: "uxprt", Hrp: "persistence", CoinType: coin.Persistence, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Akash:           {ChainID: "akashnet-2", Denom: "uakt", Hrp: "akash", CoinType: coin.Akash, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Osmosis:         {ChainID: "osmosis-1", Denom: "uosmo", Hrp: "osmo", CoinType: coin.Osmosis, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Kava:            {ChainID: "kava_2222-10", Denom: "ukava", Hrp: "kava", CoinType: coin.Kava, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Band:            {ChainID: "laozi-mainnet", Denom: "uband", Hrp: "band", CoinType: coin.Band, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Bluzelle:        {ChainID: "bluzelle-mainnet", Denom: "ubnt", Hrp: "bluzelle", CoinType: coin.Bluzelle, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Cryptoorg:       {ChainID: "crypto-org-chain-mainnet-1", Denom: "basecro", Hrp: "cro", CoinType: coin.Cryptoorg, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Secret:          {ChainID: "secret-4", Denom: "uscrt", Hrp: "secret", CoinType: coin.Secret, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Terra:           {ChainID: "columbus-5", Denom: "uluna", Hrp: "terra", CoinType: coin.Terra, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Terrav2:         {ChainID: "phoenix-1", Denom: "uluna", Hrp: "terra", CoinType: coin.Terrav2, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Agoric:          {ChainID: "agoric-3", Denom: "ubld", Hrp: "agoric", CoinType: coin.Agoric, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Dydx:            {ChainID: "dydx-mainnet-1", Denom: "adydx", Hrp: "dYdX", CoinType: coin.Dydx, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Nativeinjective: {ChainID: "injective-1", Denom: "inj", Hrp: "inj", CoinType: coin.Nativeinjective, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1Extended},
	coin.Nativecanto:     {ChainID: "canto_7700-1", Denom: "acanto", Hrp: "canto", CoinType: coin.Nativecanto, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1Extended},
	coin.Nativeevmos:     {ChainID: "evmos_9001-2", Denom: "aevmos", Hrp: "evmos", CoinType: coin.Nativeevmos, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1Extended},
	coin.Thorchain:       {ChainID: "thorchain-mainnet-v1", Denom: "rune", Hrp: "thor", CoinType: coin.Thorchain, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1},
	coin.Zetachain:       {ChainID: "zetachain_7000-1", Denom: "azeta", Hrp: "zeta", CoinType: coin.Zetachain, PublicKeyType: cosmosproto.SignerPublicKeyType_Secp256k1Extended},
}

type CosmosTransactionBuilder struct {
	coinType      coin.CoinType
	fromAddress   string
	toAddress     string
	amount        string
	denom         string
	chainID       string
	accountNumber uint64
	sequence      uint64
	fee           uint64
	gas           uint64
	memo          string
	privateKey    []byte
	publicKey     []byte
}

func NewCosmosTransaction() *CosmosTransactionBuilder {
	return &CosmosTransactionBuilder{
		gas:  200000,
		fee:  1000,
		memo: "",
	}
}

func (b *CosmosTransactionBuilder) CoinType(ct coin.CoinType) *CosmosTransactionBuilder {
	b.coinType = ct
	return b
}

func (b *CosmosTransactionBuilder) From(address string) *CosmosTransactionBuilder {
	b.fromAddress = address
	return b
}

func (b *CosmosTransactionBuilder) To(address string) *CosmosTransactionBuilder {
	b.toAddress = address
	return b
}

func (b *CosmosTransactionBuilder) Amount(amount string) *CosmosTransactionBuilder {
	b.amount = amount
	return b
}

func (b *CosmosTransactionBuilder) Denom(denom string) *CosmosTransactionBuilder {
	b.denom = denom
	return b
}

func (b *CosmosTransactionBuilder) ChainID(chainID string) *CosmosTransactionBuilder {
	b.chainID = chainID
	return b
}

func (b *CosmosTransactionBuilder) AccountNumber(num uint64) *CosmosTransactionBuilder {
	b.accountNumber = num
	return b
}

func (b *CosmosTransactionBuilder) Sequence(seq uint64) *CosmosTransactionBuilder {
	b.sequence = seq
	return b
}

func (b *CosmosTransactionBuilder) Fee(fee uint64) *CosmosTransactionBuilder {
	b.fee = fee
	return b
}

func (b *CosmosTransactionBuilder) Gas(gas uint64) *CosmosTransactionBuilder {
	b.gas = gas
	return b
}

func (b *CosmosTransactionBuilder) Memo(memo string) *CosmosTransactionBuilder {
	b.memo = memo
	return b
}

func (b *CosmosTransactionBuilder) PrivateKey(key []byte) *CosmosTransactionBuilder {
	b.privateKey = key
	return b
}

func (b *CosmosTransactionBuilder) PublicKey(key []byte) *CosmosTransactionBuilder {
	b.publicKey = key
	return b
}

func (b *CosmosTransactionBuilder) Validate() error {
	if b.toAddress == "" {
		return errors.New("recipient address is required")
	}
	if b.fromAddress == "" {
		return errors.New("sender address is required")
	}
	if b.amount == "" {
		return errors.New("amount is required")
	}
	if b.denom == "" {
		return errors.New("denom is required")
	}
	if b.chainID == "" {
		return errors.New("chain ID is required")
	}
	if len(b.privateKey) == 0 {
		return errors.New("private key is required")
	}
	return nil
}

func (b *CosmosTransactionBuilder) Sign() ([]byte, error) {
	result, err := b.SignWithResult()
	if err != nil {
		return nil, err
	}
	return []byte(result.Serialized), nil
}

func (b *CosmosTransactionBuilder) SignWithResult() (*CosmosTxResult, error) {
	if err := b.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	sendMsg := &cosmosproto.Message_Send{
		FromAddress: b.fromAddress,
		ToAddress:   b.toAddress,
		Amounts: []*cosmosproto.Amount{
			{
				Denom:  b.denom,
				Amount: b.amount,
			},
		},
	}

	msg := &cosmosproto.Message{
		MessageOneof: &cosmosproto.Message_SendCoinsMessage{
			SendCoinsMessage: sendMsg,
		},
	}

	input := &cosmosproto.SigningInput{
		SigningMode:   cosmosproto.SigningMode_Protobuf,
		ChainId:       b.chainID,
		AccountNumber: b.accountNumber,
		Sequence:      b.sequence,
		Memo:          b.memo,
		PrivateKey:    b.privateKey,
		PublicKey:     b.publicKey,
		Fee: &cosmosproto.Fee{
			Amounts: []*cosmosproto.Amount{
				{
					Denom:  b.denom,
					Amount: fmt.Sprintf("%d", b.fee),
				},
			},
			Gas: b.gas,
		},
		Messages: []*cosmosproto.Message{msg},
		Mode:     cosmosproto.BroadcastMode_SYNC,
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

	var output cosmosproto.SigningOutput
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

	serialized := output.GetSerialized()
	if serialized == "" {
		json := output.GetJson()
		if json == "" {
			return nil, errors.New("signed transaction is empty")
		}
		return &CosmosTxResult{
			Serialized: json,
			Json:       json,
			Signature:  output.GetSignature(),
		}, nil
	}

	return &CosmosTxResult{
		Serialized:    serialized,
		Json:          output.GetJson(),
		Signature:     output.GetSignature(),
		SignatureJson: output.GetSignatureJson(),
	}, nil
}

type CosmosTxResult struct {
	Serialized    string
	Json          string
	Signature     []byte
	SignatureJson string
}

func SignCosmosTransaction(
	coinType coin.CoinType,
	privateKey []byte,
	publicKey []byte,
	fromAddress string,
	toAddress string,
	amount string,
	accountNumber uint64,
	sequence uint64,
) ([]byte, error) {
	config, ok := CosmosChainConfigs[coinType]
	if !ok {
		return nil, fmt.Errorf("unsupported Cosmos chain: %d", coinType)
	}

	result, err := NewCosmosTransaction().
		CoinType(coinType).
		From(fromAddress).
		To(toAddress).
		Amount(amount).
		Denom(config.Denom).
		ChainID(config.ChainID).
		AccountNumber(accountNumber).
		Sequence(sequence).
		PrivateKey(privateKey).
		PublicKey(publicKey).
		SignWithResult()

	if err != nil {
		return nil, err
	}
	return []byte(result.Serialized), nil
}
