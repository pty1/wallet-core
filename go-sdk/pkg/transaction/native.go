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
	"unsafe"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	aeternityproto "github.com/trustwallet/go-wallet-core/pkg/proto/aeternity"
	aionproto "github.com/trustwallet/go-wallet-core/pkg/proto/aion"
	algorandproto "github.com/trustwallet/go-wallet-core/pkg/proto/algorand"
	aptosproto "github.com/trustwallet/go-wallet-core/pkg/proto/aptos"
	binanceproto "github.com/trustwallet/go-wallet-core/pkg/proto/binance"
	cardanoproto "github.com/trustwallet/go-wallet-core/pkg/proto/cardano"
	"github.com/trustwallet/go-wallet-core/pkg/proto/common"
	eosproto "github.com/trustwallet/go-wallet-core/pkg/proto/eos"
	everscaleproto "github.com/trustwallet/go-wallet-core/pkg/proto/everscale"
	filecoinproto "github.com/trustwallet/go-wallet-core/pkg/proto/filecoin"
	fioproto "github.com/trustwallet/go-wallet-core/pkg/proto/fio"
	greenfieldproto "github.com/trustwallet/go-wallet-core/pkg/proto/greenfield"
	hederaproto "github.com/trustwallet/go-wallet-core/pkg/proto/hedera"
	iconproto "github.com/trustwallet/go-wallet-core/pkg/proto/icon"
	internetcomputerproto "github.com/trustwallet/go-wallet-core/pkg/proto/internetcomputer"
	iostproto "github.com/trustwallet/go-wallet-core/pkg/proto/iost"
	iotexproto "github.com/trustwallet/go-wallet-core/pkg/proto/iotex"
	multiversxproto "github.com/trustwallet/go-wallet-core/pkg/proto/multiversx"
	nanoproto "github.com/trustwallet/go-wallet-core/pkg/proto/nano"
	nearproto "github.com/trustwallet/go-wallet-core/pkg/proto/near"
	nebulasproto "github.com/trustwallet/go-wallet-core/pkg/proto/nebulas"
	neoproto "github.com/trustwallet/go-wallet-core/pkg/proto/neo"
	nervosproto "github.com/trustwallet/go-wallet-core/pkg/proto/nervos"
	nimiqproto "github.com/trustwallet/go-wallet-core/pkg/proto/nimiq"
	nulsproto "github.com/trustwallet/go-wallet-core/pkg/proto/nuls"
	ontologyproto "github.com/trustwallet/go-wallet-core/pkg/proto/ontology"
	pactusproto "github.com/trustwallet/go-wallet-core/pkg/proto/pactus"
	polkadotproto "github.com/trustwallet/go-wallet-core/pkg/proto/polkadot"
	polymeshproto "github.com/trustwallet/go-wallet-core/pkg/proto/polymesh"
	rippleproto "github.com/trustwallet/go-wallet-core/pkg/proto/ripple"
	solanaproto "github.com/trustwallet/go-wallet-core/pkg/proto/solana"
	stellarproto "github.com/trustwallet/go-wallet-core/pkg/proto/stellar"
	suiproto "github.com/trustwallet/go-wallet-core/pkg/proto/sui"
	tezosproto "github.com/trustwallet/go-wallet-core/pkg/proto/tezos"
	tonproto "github.com/trustwallet/go-wallet-core/pkg/proto/theopennetwork"
	thetaproto "github.com/trustwallet/go-wallet-core/pkg/proto/theta"
	tronproto "github.com/trustwallet/go-wallet-core/pkg/proto/tron"
	vechainproto "github.com/trustwallet/go-wallet-core/pkg/proto/vechain"
	wavesproto "github.com/trustwallet/go-wallet-core/pkg/proto/waves"
	zilliqaeproto "github.com/trustwallet/go-wallet-core/pkg/proto/zilliqa"
	"google.golang.org/protobuf/proto"
)

func signTransactionRaw(coinType coin.CoinType, input proto.Message, output proto.Message) error {
	inputBytes, err := proto.Marshal(input)
	if err != nil {
		return fmt.Errorf("failed to marshal input: %w", err)
	}

	cInputData := C.TWDataCreateWithBytes(
		(*C.uint8_t)(&inputBytes[0]),
		C.size_t(len(inputBytes)),
	)
	defer C.TWDataDelete(cInputData)

	cOutputData := C.TWAnySignerSign(cInputData, C.enum_TWCoinType(coinType))
	defer C.TWDataDelete(cOutputData)

	outputSize := C.TWDataSize(cOutputData)
	if outputSize == 0 {
		return errors.New("signing returned empty result")
	}

	outputBytes := make([]byte, outputSize)
	C.memcpy(
		unsafe.Pointer(&outputBytes[0]),
		unsafe.Pointer(C.TWDataBytes(cOutputData)),
		C.size_t(outputSize),
	)

	if err := proto.Unmarshal(outputBytes, output); err != nil {
		return fmt.Errorf("failed to unmarshal signing output: %w", err)
	}

	return nil
}

func extractSignedTransaction(output proto.Message) ([]byte, error) {
	if o, ok := output.(interface {
		GetError() common.SigningError
		GetErrorMessage() string
	}); ok {
		if o.GetError() != common.SigningError_OK {
			return nil, fmt.Errorf("signing error: %s", o.GetErrorMessage())
		}
	}

	switch o := output.(type) {
	case *solanaproto.SigningOutput:
		return []byte(o.Encoded), nil
	case *tronproto.SigningOutput:
		return []byte(o.Json), nil
	case *rippleproto.SigningOutput:
		return o.Encoded, nil
	case *stellarproto.SigningOutput:
		return []byte(o.Signature), nil
	case *polkadotproto.SigningOutput:
		return o.Encoded, nil
	case *cardanoproto.SigningOutput:
		return o.Encoded, nil
	case *tezosproto.SigningOutput:
		return o.Encoded, nil
	case *eosproto.SigningOutput:
		return []byte(o.JsonEncoded), nil
	case *nearproto.SigningOutput:
		return o.SignedTransaction, nil
	case *filecoinproto.SigningOutput:
		return []byte(o.Json), nil
	case *algorandproto.SigningOutput:
		return o.Encoded, nil
	case *aptosproto.SigningOutput:
		return o.Encoded, nil
	case *suiproto.SigningOutput:
		return []byte(o.Signature), nil
	case *binanceproto.SigningOutput:
		return o.Encoded, nil
	case *fioproto.SigningOutput:
		return []byte(o.Json), nil
	case *vechainproto.SigningOutput:
		return o.Encoded, nil
	case *wavesproto.SigningOutput:
		return []byte(o.Json), nil
	case *zilliqaeproto.SigningOutput:
		return []byte(o.Json), nil
	case *nervosproto.SigningOutput:
		return []byte(o.TransactionJson), nil
	case *ontologyproto.SigningOutput:
		return o.Encoded, nil
	case *multiversxproto.SigningOutput:
		return []byte(o.Encoded), nil
	case *tonproto.SigningOutput:
		return []byte(o.Encoded), nil
	case *everscaleproto.SigningOutput:
		return []byte(o.Encoded), nil
	case *pactusproto.SigningOutput:
		return o.SignedTransactionData, nil
	case *polymeshproto.SigningOutput:
		return o.Encoded, nil
	case *hederaproto.SigningOutput:
		return o.Encoded, nil
	case *iotexproto.SigningOutput:
		return []byte(o.Encoded), nil
	case *nanoproto.SigningOutput:
		return []byte(o.Json), nil
	case *aeternityproto.SigningOutput:
		return []byte(o.Encoded), nil
	case *aionproto.SigningOutput:
		return o.Encoded, nil
	case *iconproto.SigningOutput:
		return []byte(o.Encoded), nil
	case *iostproto.SigningOutput:
		return o.Encoded, nil
	case *internetcomputerproto.SigningOutput:
		return o.SignedTransaction, nil
	case *nebulasproto.SigningOutput:
		return []byte(o.Raw), nil
	case *neoproto.SigningOutput:
		return o.Encoded, nil
	case *nimiqproto.SigningOutput:
		return o.Encoded, nil
	case *nulsproto.SigningOutput:
		return o.Encoded, nil
	case *thetaproto.SigningOutput:
		return o.Encoded, nil
	case *greenfieldproto.SigningOutput:
		return []byte(o.Serialized), nil
	default:
		return nil, fmt.Errorf("unsupported output type: %T", output)
	}
}

func SignTransaction(coinType coin.CoinType, input proto.Message, output proto.Message) ([]byte, error) {
	if err := signTransactionRaw(coinType, input, output); err != nil {
		return nil, err
	}
	return extractSignedTransaction(output)
}

// BuildSolanaTransaction builds a Solana transfer transaction
func BuildSolanaTransaction(privateKey []byte, fromAddress, toAddress string, amount uint64, recentBlockhash string) *solanaproto.SigningInput {
	return &solanaproto.SigningInput{
		PrivateKey:      privateKey,
		RecentBlockhash: recentBlockhash,
		Sender:          fromAddress,
		TransactionType: &solanaproto.SigningInput_TransferTransaction{
			TransferTransaction: &solanaproto.Transfer{
				Recipient: toAddress,
				Value:     amount,
			},
		},
	}
}

// BuildTronTransaction builds a Tron transfer transaction
func BuildTronTransaction(privateKey []byte, fromAddress, toAddress string, amount int64, timestamp int64, blockTimestamp int64, blockNumber int64) *tronproto.SigningInput {
	return &tronproto.SigningInput{
		PrivateKey: privateKey,
		Transaction: &tronproto.Transaction{
			Timestamp: timestamp,
			BlockHeader: &tronproto.BlockHeader{
				Timestamp: blockTimestamp,
				Number:    blockNumber,
			},
			ContractOneof: &tronproto.Transaction_Transfer{
				Transfer: &tronproto.TransferContract{
					OwnerAddress: fromAddress,
					ToAddress:    toAddress,
					Amount:       amount,
				},
			},
		},
	}
}

// BuildXRPTransaction builds an XRP transfer transaction
func BuildXRPTransaction(privateKey []byte, account, destination string, amount int64, sequence uint32, fee int64, lastLedgerSeq uint32) *rippleproto.SigningInput {
	return &rippleproto.SigningInput{
		PrivateKey:         privateKey,
		Account:            account,
		Sequence:           sequence,
		Fee:                fee,
		LastLedgerSequence: lastLedgerSeq,
		OperationOneof: &rippleproto.SigningInput_OpPayment{
			OpPayment: &rippleproto.OperationPayment{
				Destination: destination,
				AmountOneof: &rippleproto.OperationPayment_Amount{
					Amount: amount,
				},
			},
		},
	}
}

// BuildStellarTransaction builds a Stellar transfer transaction
func BuildStellarTransaction(privateKey []byte, account, destination string, amount int64, sequence int64, fee int32, passphrase string) *stellarproto.SigningInput {
	return &stellarproto.SigningInput{
		PrivateKey: privateKey,
		Account:    account,
		Sequence:   sequence,
		Fee:        fee,
		Passphrase: passphrase,
		OperationOneof: &stellarproto.SigningInput_OpPayment{
			OpPayment: &stellarproto.OperationPayment{
				Destination: destination,
				Amount:      amount,
			},
		},
	}
}

// BuildKinTransaction builds a Kin transfer transaction
func BuildKinTransaction(privateKey []byte, account, destination string, amount int64, sequence int64, fee int32) *stellarproto.SigningInput {
	return &stellarproto.SigningInput{
		PrivateKey: privateKey,
		Account:    account,
		Sequence:   sequence,
		Fee:        fee,
		Passphrase: "Kin Mainnet ; December 2018",
		OperationOneof: &stellarproto.SigningInput_OpPayment{
			OpPayment: &stellarproto.OperationPayment{
				Destination: destination,
				Amount:      amount,
			},
		},
	}
}

// BuildPolkadotTransaction builds a Polkadot transfer transaction
func BuildPolkadotTransaction(privateKey []byte, toAddress string, value []byte, blockHash, genesisHash []byte, nonce uint64, specVersion, txVersion uint32) *polkadotproto.SigningInput {
	return &polkadotproto.SigningInput{
		PrivateKey:         privateKey,
		BlockHash:          blockHash,
		GenesisHash:        genesisHash,
		Nonce:              nonce,
		SpecVersion:        specVersion,
		TransactionVersion: txVersion,
		MessageOneof: &polkadotproto.SigningInput_BalanceCall{
			BalanceCall: &polkadotproto.Balance{
				MessageOneof: &polkadotproto.Balance_Transfer_{
					Transfer: &polkadotproto.Balance_Transfer{
						ToAddress: toAddress,
						Value:     value,
					},
				},
			},
		},
	}
}

// BuildKusamaTransaction builds a Kusama transfer transaction
func BuildKusamaTransaction(privateKey []byte, toAddress string, value []byte, blockHash, genesisHash []byte, nonce uint64, specVersion, txVersion uint32) *polkadotproto.SigningInput {
	return &polkadotproto.SigningInput{
		PrivateKey:         privateKey,
		BlockHash:          blockHash,
		GenesisHash:        genesisHash,
		Nonce:              nonce,
		SpecVersion:        specVersion,
		TransactionVersion: txVersion,
		Network:            2,
		MessageOneof: &polkadotproto.SigningInput_BalanceCall{
			BalanceCall: &polkadotproto.Balance{
				MessageOneof: &polkadotproto.Balance_Transfer_{
					Transfer: &polkadotproto.Balance_Transfer{
						ToAddress: toAddress,
						Value:     value,
					},
				},
			},
		},
	}
}

// BuildCardanoTransaction builds a Cardano transfer transaction
func BuildCardanoTransaction(privateKey [][]byte, utxos []*cardanoproto.TxInput, transferMsg *cardanoproto.Transfer, ttl uint64) *cardanoproto.SigningInput {
	return &cardanoproto.SigningInput{
		PrivateKey:      privateKey,
		Utxos:           utxos,
		TransferMessage: transferMsg,
		Ttl:             ttl,
	}
}

// BuildTezosTransaction builds a Tezos transaction
func BuildTezosTransaction(privateKey []byte, operations *tezosproto.OperationList) *tezosproto.SigningInput {
	return &tezosproto.SigningInput{
		PrivateKey:    privateKey,
		OperationList: operations,
	}
}

// BuildEOSTransaction builds an EOS transfer transaction
func BuildEOSTransaction(privateKey []byte, fromAddress, toAddress string, amount int64, chainID, refBlockID []byte, refBlockTime int32) *eosproto.SigningInput {
	return &eosproto.SigningInput{
		PrivateKey:         privateKey,
		PrivateKeyType:     eosproto.KeyType_MODERNK1,
		ChainId:            chainID,
		ReferenceBlockId:   refBlockID,
		ReferenceBlockTime: refBlockTime,
		Currency:           "eosio.token",
		Sender:             fromAddress,
		Recipient:          toAddress,
		Asset: &eosproto.Asset{
			Amount:   amount,
			Decimals: 4,
			Symbol:   "EOS",
		},
	}
}

// BuildWAXTransaction builds a WAX transfer transaction
func BuildWAXTransaction(privateKey []byte, fromAddress, toAddress string, amount int64, chainID, refBlockID []byte, refBlockTime int32) *eosproto.SigningInput {
	return &eosproto.SigningInput{
		PrivateKey:         privateKey,
		PrivateKeyType:     eosproto.KeyType_MODERNK1,
		ChainId:            chainID,
		ReferenceBlockId:   refBlockID,
		ReferenceBlockTime: refBlockTime,
		Currency:           "eosio.token",
		Sender:             fromAddress,
		Recipient:          toAddress,
		Asset: &eosproto.Asset{
			Amount:   amount,
			Decimals: 8,
			Symbol:   "WAX",
		},
	}
}

// NEAR Transaction Signing
// BuildNEARTransaction builds a NEAR transfer transaction
func BuildNEARTransaction(privateKey []byte, signerID, receiverID string, amount []byte, nonce uint64, blockHash []byte) *nearproto.SigningInput {
	return &nearproto.SigningInput{
		PrivateKey: privateKey,
		SignerId:   signerID,
		ReceiverId: receiverID,
		Nonce:      nonce,
		BlockHash:  blockHash,
		Actions: []*nearproto.Action{
			{
				Payload: &nearproto.Action_Transfer{
					Transfer: &nearproto.Transfer{
						Deposit: amount,
					},
				},
			},
		},
	}
}

// BuildFilecoinTransaction builds a Filecoin transfer transaction
func BuildFilecoinTransaction(privateKey, publicKey []byte, toAddress string, amount []byte, nonce uint64, gasLimit int64, gasFeeCap, gasPremium []byte) *filecoinproto.SigningInput {
	return &filecoinproto.SigningInput{
		PrivateKey: privateKey,
		PublicKey:  publicKey,
		To:         toAddress,
		Value:      amount,
		Nonce:      nonce,
		GasLimit:   gasLimit,
		GasFeeCap:  gasFeeCap,
		GasPremium: gasPremium,
	}
}

// BuildAlgorandTransaction builds an Algorand transfer transaction
func BuildAlgorandTransaction(privateKey, publicKey []byte, toAddress string, amount uint64, fee uint64, firstRound, lastRound uint64, genesisID string, genesisHash []byte) *algorandproto.SigningInput {
	return &algorandproto.SigningInput{
		PrivateKey:  privateKey,
		PublicKey:   publicKey,
		GenesisId:   genesisID,
		GenesisHash: genesisHash,
		Fee:         fee,
		FirstRound:  firstRound,
		LastRound:   lastRound,
		MessageOneof: &algorandproto.SigningInput_Transfer{
			Transfer: &algorandproto.Transfer{
				ToAddress: toAddress,
				Amount:    amount,
			},
		},
	}
}

// BuildAptosTransaction builds an Aptos transfer transaction
func BuildAptosTransaction(privateKey []byte, fromAddress, toAddress string, amount uint64, sequenceNumber int64, maxGasAmount uint64, gasUnitPrice uint64, expirationTimestampSecs uint64, chainID uint32) *aptosproto.SigningInput {
	return &aptosproto.SigningInput{
		PrivateKey:              privateKey,
		Sender:                  fromAddress,
		SequenceNumber:          sequenceNumber,
		MaxGasAmount:            maxGasAmount,
		GasUnitPrice:            gasUnitPrice,
		ExpirationTimestampSecs: expirationTimestampSecs,
		ChainId:                 chainID,
		TransactionPayload: &aptosproto.SigningInput_Transfer{
			Transfer: &aptosproto.TransferMessage{
				To:     toAddress,
				Amount: amount,
			},
		},
	}
}

// BuildSuiTransaction builds a Sui transfer transaction
func BuildSuiTransaction(privateKey []byte, signer string, pay *suiproto.Pay, gasBudget, referenceGasPrice uint64) *suiproto.SigningInput {
	return &suiproto.SigningInput{
		PrivateKey:        privateKey,
		Signer:            signer,
		GasBudget:         gasBudget,
		ReferenceGasPrice: referenceGasPrice,
		TransactionPayload: &suiproto.SigningInput_Pay{
			Pay: pay,
		},
	}
}

// BuildBinanceChainTransaction builds a Binance Chain transfer transaction
func BuildBinanceChainTransaction(privateKey []byte, fromAddress, toAddress string, amount int64, denom string, accountNumber, sequence int64, chainID string) *binanceproto.SigningInput {
	return &binanceproto.SigningInput{
		PrivateKey:    privateKey,
		AccountNumber: accountNumber,
		Sequence:      sequence,
		ChainId:       chainID,
		OrderOneof: &binanceproto.SigningInput_SendOrder{
			SendOrder: &binanceproto.SendOrder{
				Inputs: []*binanceproto.SendOrder_Input{
					{
						Address: []byte(fromAddress),
						Coins: []*binanceproto.SendOrder_Token{
							{Denom: denom, Amount: amount},
						},
					},
				},
				Outputs: []*binanceproto.SendOrder_Output{
					{
						Address: []byte(toAddress),
						Coins: []*binanceproto.SendOrder_Token{
							{Denom: denom, Amount: amount},
						},
					},
				},
			},
		},
	}
}

// FIO Transaction Signing
// BuildFIOTransaction builds a FIO transaction
func BuildFIOTransaction(privateKey []byte, toAddress string, amount int64, fee int64) *fioproto.SigningInput {
	return &fioproto.SigningInput{
		PrivateKey: privateKey,
		Action: &fioproto.Action{
			MessageOneof: &fioproto.Action_TransferMessage{
				TransferMessage: &fioproto.Action_Transfer{
					PayeePublicKey: toAddress,
					Amount:         uint64(amount),
					Fee:            uint64(fee),
				},
			},
		},
		Tpid: "fee@fio",
	}

}

// BuildHederaTransaction builds a Hedera transfer transaction
func BuildHederaTransaction(privateKey []byte, fromAddress, toAddress string, amount int64, fee uint64, nodeAccountID string) *hederaproto.SigningInput {
	return &hederaproto.SigningInput{
		PrivateKey: privateKey,
		Body: &hederaproto.TransactionBody{
			TransactionID: &hederaproto.TransactionID{
				AccountID:             fromAddress,
				TransactionValidStart: &hederaproto.Timestamp{Seconds: 0, Nanos: 0},
			},
			NodeAccountID:            nodeAccountID,
			TransactionFee:           fee,
			TransactionValidDuration: 120,
			Data: &hederaproto.TransactionBody_Transfer{
				Transfer: &hederaproto.TransferMessage{
					From:   fromAddress,
					To:     toAddress,
					Amount: amount,
				},
			},
		},
	}
}

// VeChain Transaction Signing
// BuildVeChainTransaction builds a VeChain transaction
func BuildVeChainTransaction(privateKey []byte, toAddress string, amount []byte, nonce uint64, gasLimit uint64, gasPriceCoef uint32, chainTag uint32, blockRef uint64) *vechainproto.SigningInput {
	return &vechainproto.SigningInput{
		PrivateKey: privateKey,
		Nonce:      nonce,
		Gas:        gasLimit,
		Clauses: []*vechainproto.Clause{
			{To: toAddress, Value: amount},
		},
		GasPriceCoef: gasPriceCoef,
		ChainTag:     chainTag,
		BlockRef:     blockRef,
	}
}

// Waves Transaction Signing
// BuildWavesTransaction builds a Waves transaction
func BuildWavesTransaction(privateKey []byte, toAddress string, amount int64, fee int64, timestamp int64) *wavesproto.SigningInput {
	return &wavesproto.SigningInput{
		PrivateKey: privateKey,
		Timestamp:  timestamp,
		MessageOneof: &wavesproto.SigningInput_TransferMessage{
			TransferMessage: &wavesproto.TransferMessage{
				Amount:   amount,
				To:       toAddress,
				Fee:      fee,
				Asset:    "WAVES",
				FeeAsset: "WAVES",
			},
		},
	}
}

// Zilliqa Transaction Signing
// BuildZilliqaTransaction builds a Zilliqa transaction
func BuildZilliqaTransaction(privateKey []byte, toAddress string, amount []byte, nonce uint64, gasLimit uint64, gasPrice []byte, version uint32) *zilliqaeproto.SigningInput {
	return &zilliqaeproto.SigningInput{
		PrivateKey: privateKey,
		To:         toAddress,
		Nonce:      nonce,
		GasLimit:   gasLimit,
		GasPrice:   gasPrice,
		Version:    version,
		Transaction: &zilliqaeproto.Transaction{
			MessageOneof: &zilliqaeproto.Transaction_Transfer_{
				Transfer: &zilliqaeproto.Transaction_Transfer{
					Amount: amount,
				},
			},
		},
	}
}

// Nervos Transaction Signing
// BuildNervosTransaction builds a Nervos transaction
func BuildNervosTransaction(privateKey [][]byte, toAddress, changeAddress string, amount uint64, cells []*nervosproto.Cell) *nervosproto.SigningInput {
	return &nervosproto.SigningInput{
		PrivateKey: privateKey,
		Cell:       cells,
		OperationOneof: &nervosproto.SigningInput_NativeTransfer{
			NativeTransfer: &nervosproto.NativeTransfer{
				ToAddress:     toAddress,
				ChangeAddress: changeAddress,
				Amount:        amount,
			},
		},
	}
}

// Ontology Transaction Signing
// BuildOntologyTransaction builds a Ontology transaction
func BuildOntologyTransaction(ownerPrivateKey []byte, ownerAddress, toAddress string, amount uint64, gasPrice, gasLimit uint64, nonce uint32) *ontologyproto.SigningInput {
	return &ontologyproto.SigningInput{
		Contract:        "ONT",
		Method:          "transfer",
		OwnerPrivateKey: ownerPrivateKey,
		PayerPrivateKey: ownerPrivateKey,
		OwnerAddress:    ownerAddress,
		ToAddress:       toAddress,
		Amount:          amount,
		GasPrice:        gasPrice,
		GasLimit:        gasLimit,
		Nonce:           nonce,
	}
}

// MultiversX Transaction Signing
// BuildMultiversXTransaction builds a MultiversX transaction
func BuildMultiversXTransaction(privateKey []byte, senderAddress, receiverAddress, amount string, nonce uint64, chainID string, gasLimit, gasPrice uint64) *multiversxproto.SigningInput {
	return &multiversxproto.SigningInput{
		PrivateKey: privateKey,
		ChainId:    chainID,
		GasPrice:   gasPrice,
		GasLimit:   gasLimit,
		MessageOneof: &multiversxproto.SigningInput_EgldTransfer{
			EgldTransfer: &multiversxproto.EGLDTransfer{
				Accounts: &multiversxproto.Accounts{
					SenderNonce: nonce,
					Sender:      senderAddress,
					Receiver:    receiverAddress,
				},
				Amount: amount,
			},
		},
	}
}

// TON Transaction Signing
// BuildTONTransaction builds a TON transaction
func BuildTONTransaction(privateKey []byte, toAddress string, amount []byte, sequenceNumber uint32, expireAt uint32) *tonproto.SigningInput {
	return &tonproto.SigningInput{
		PrivateKey:     privateKey,
		SequenceNumber: sequenceNumber,
		ExpireAt:       expireAt,
		WalletVersion:  tonproto.WalletVersion_WALLET_V4_R2,
		Messages: []*tonproto.Transfer{
			{
				Dest:   toAddress,
				Amount: amount,
				Mode:   3,
			},
		},
	}
}

// Everscale Transaction Signing
// BuildEverscaleTransaction builds a Everscale transaction
func BuildEverscaleTransaction(privateKey []byte, toAddress string, amount uint64, expiredAt uint32) *everscaleproto.SigningInput {
	return &everscaleproto.SigningInput{
		PrivateKey: privateKey,
		ActionOneof: &everscaleproto.SigningInput_Transfer{
			Transfer: &everscaleproto.Transfer{
				To:        toAddress,
				Amount:    amount,
				ExpiredAt: expiredAt,
				Behavior:  everscaleproto.MessageBehavior_SimpleTransfer,
				Bounce:    true,
			},
		},
	}
}

// Pactus Transaction Signing
// BuildPactusTransaction builds a Pactus transaction
func BuildPactusTransaction(privateKey []byte, fromAddress, toAddress string, amount int64, fee int64, lockTime uint32) *pactusproto.SigningInput {
	return &pactusproto.SigningInput{
		PrivateKey: privateKey,
		Transaction: &pactusproto.TransactionMessage{
			LockTime: lockTime,
			Fee:      fee,
			Memo:     "",
			Payload: &pactusproto.TransactionMessage_Transfer{
				Transfer: &pactusproto.TransferPayload{
					Sender:   fromAddress,
					Receiver: toAddress,
					Amount:   amount,
				},
			},
		},
	}
}

// Polymesh Transaction Signing
// BuildPolymeshTransaction builds a Polymesh transaction
func BuildPolymeshTransaction(privateKey []byte, toAddress string, amount []byte, nonce uint64, blockHash []byte, genesisHash []byte, specVersion, transactionVersion uint32) *polymeshproto.SigningInput {
	return &polymeshproto.SigningInput{
		PrivateKey:         privateKey,
		Nonce:              nonce,
		BlockHash:          blockHash,
		GenesisHash:        genesisHash,
		SpecVersion:        specVersion,
		TransactionVersion: transactionVersion,
		RuntimeCall: &polymeshproto.RuntimeCall{
			PalletOneof: &polymeshproto.RuntimeCall_BalanceCall{
				BalanceCall: &polymeshproto.Balance{
					MessageOneof: &polymeshproto.Balance_Transfer_{
						Transfer: &polymeshproto.Balance_Transfer{
							ToAddress: toAddress,
							Value:     amount,
						},
					},
				},
			},
		},
	}
}

// BuildGreenfieldTransaction builds a Greenfield transaction using Cosmos builder
func BuildGreenfieldTransaction(privateKey []byte, fromAddress, toAddress string, amount string, accountNumber, sequence uint64) *CosmosTransactionBuilder {
	return NewCosmosTransaction().
		CoinType(coin.Greenfield).
		From(fromAddress).
		To(toAddress).
		Amount(amount).
		Denom("abnbs").
		ChainID("greenfield_9000-121").
		AccountNumber(accountNumber).
		Sequence(sequence).
		PrivateKey(privateKey)
}

// IoTeX Transaction Signing
// BuildIoTeXTransaction builds a IoTeX transaction
func BuildIoTeXTransaction(privateKey []byte, toAddress, amount string, nonce uint64, gasLimit uint64, gasPrice string, chainID uint32) *iotexproto.SigningInput {
	return &iotexproto.SigningInput{
		Version:    1,
		Nonce:      nonce,
		GasLimit:   gasLimit,
		GasPrice:   gasPrice,
		ChainID:    chainID,
		PrivateKey: privateKey,
		Action: &iotexproto.SigningInput_Transfer{
			Transfer: &iotexproto.Transfer{
				Amount:    amount,
				Recipient: toAddress,
			},
		},
	}
}

// Nano Transaction Signing
// BuildNanoTransaction builds a Nano transaction
func BuildNanoTransaction(privateKey, publicKey []byte, toAddress, representative, balance, work string) *nanoproto.SigningInput {
	return &nanoproto.SigningInput{
		PrivateKey:     privateKey,
		PublicKey:      publicKey,
		LinkOneof:      &nanoproto.SigningInput_LinkRecipient{LinkRecipient: toAddress},
		Representative: representative,
		Balance:        balance,
		Work:           work,
	}
}

// Aeternity Transaction Signing
// BuildAeternityTransaction builds a Aeternity transaction
func BuildAeternityTransaction(privateKey []byte, fromAddress, toAddress string, amount, fee []byte, ttl, nonce uint64) *aeternityproto.SigningInput {
	return &aeternityproto.SigningInput{
		PrivateKey:  privateKey,
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
		Fee:         fee,
		Ttl:         ttl,
		Nonce:       nonce,
		Payload:     "",
	}
}

// Aion Transaction Signing
// BuildAionTransaction builds a Aion transaction
func BuildAionTransaction(privateKey []byte, toAddress string, amount, nonce, gasPrice, gasLimit []byte, timestamp uint64) *aionproto.SigningInput {
	return &aionproto.SigningInput{
		PrivateKey: privateKey,
		Nonce:      nonce,
		GasPrice:   gasPrice,
		GasLimit:   gasLimit,
		ToAddress:  toAddress,
		Amount:     amount,
		Payload:    []byte{},
		Timestamp:  timestamp,
	}
}

// ICON Transaction Signing
// BuildICONTransaction builds a ICON transaction
func BuildICONTransaction(privateKey []byte, fromAddress, toAddress string, value, stepLimit, nonce, networkId []byte, timestamp int64) *iconproto.SigningInput {
	return &iconproto.SigningInput{
		PrivateKey:  privateKey,
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Value:       value,
		StepLimit:   stepLimit,
		Timestamp:   timestamp,
		Nonce:       nonce,
		NetworkId:   networkId,
	}
}

// IOST Transaction Signing
// BuildIOSTTransaction builds a IOST transaction
func BuildIOSTTransaction(accountName string, activeKey []byte, toAddress, amount string) *iostproto.SigningInput {
	return &iostproto.SigningInput{
		Account: &iostproto.AccountInfo{
			Name:      accountName,
			ActiveKey: activeKey,
		},
		TransferDestination: toAddress,
		TransferAmount:      amount,
		TransferMemo:        "",
	}
}

// Internet Computer Transaction Signing
// BuildInternetComputerTransaction builds a InternetComputer transaction
func BuildInternetComputerTransaction(privateKey []byte, toAccountIdentifier string, amount, memo, currentTimestampNanos, permittedDrift uint64) *internetcomputerproto.SigningInput {
	return &internetcomputerproto.SigningInput{
		PrivateKey: privateKey,
		Transaction: &internetcomputerproto.Transaction{
			TransactionOneof: &internetcomputerproto.Transaction_Transfer_{
				Transfer: &internetcomputerproto.Transaction_Transfer{
					ToAccountIdentifier:   toAccountIdentifier,
					Amount:                amount,
					Memo:                  memo,
					CurrentTimestampNanos: currentTimestampNanos,
					PermittedDrift:        permittedDrift,
				},
			},
		},
	}
}

// Nebulas Transaction Signing
// BuildNebulasTransaction builds a Nebulas transaction
func BuildNebulasTransaction(privateKey []byte, fromAddress, toAddress string, amount, nonce, gasPrice, gasLimit, chainId, timestamp []byte) *nebulasproto.SigningInput {
	return &nebulasproto.SigningInput{
		PrivateKey:  privateKey,
		FromAddress: fromAddress,
		ChainId:     chainId,
		Nonce:       nonce,
		GasPrice:    gasPrice,
		GasLimit:    gasLimit,
		ToAddress:   toAddress,
		Amount:      amount,
		Timestamp:   timestamp,
		Payload:     "",
	}
}

// NEO Transaction Signing
// BuildNEOTransaction builds a NEO transaction
func BuildNEOTransaction(privateKey []byte, inputs []*neoproto.TransactionInput, outputs []*neoproto.TransactionOutput) *neoproto.SigningInput {
	return &neoproto.SigningInput{
		PrivateKey: privateKey,
		Inputs:     inputs,
		Outputs:    outputs,
	}

}

// Nimiq Transaction Signing
// BuildNimiqTransaction builds a Nimiq transaction
func BuildNimiqTransaction(privateKey []byte, destination string, value, fee uint64, validityStartHeight uint32) *nimiqproto.SigningInput {
	return &nimiqproto.SigningInput{
		PrivateKey:          privateKey,
		Destination:         destination,
		Value:               value,
		Fee:                 fee,
		ValidityStartHeight: validityStartHeight,
		NetworkId:           nimiqproto.NetworkId_Mainnet,
	}
}

// NULS Transaction Signing
// BuildNULSTransaction builds a NULS transaction
func BuildNULSTransaction(privateKey []byte, from, to string, amount []byte, chainId, idassetsId uint32, nonce []byte, balance []byte, timestamp uint32) *nulsproto.SigningInput {
	return &nulsproto.SigningInput{
		PrivateKey: privateKey,
		From:       from,
		To:         to,
		Amount:     amount,
		ChainId:    chainId,
		IdassetsId: idassetsId,
		Nonce:      nonce,
		Balance:    balance,
		Timestamp:  timestamp,
	}
}

// Theta Transaction Signing
// BuildThetaTransaction builds a Theta transaction
func BuildThetaTransaction(privateKey, publicKey []byte, toAddress string, thetaAmount, tfuelAmount []byte, sequence uint64, fee []byte) *thetaproto.SigningInput {
	return &thetaproto.SigningInput{
		ChainId:     "mainnet",
		ToAddress:   toAddress,
		ThetaAmount: thetaAmount,
		TfuelAmount: tfuelAmount,
		Sequence:    sequence,
		Fee:         fee,
		PrivateKey:  privateKey,
		PublicKey:   publicKey,
	}
}

// Helper functions
func mustDecodeHex(s string) []byte {
	if len(s) > 2 && s[:2] == "0x" {
		s = s[2:]
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return []byte{}
	}
	return b
}

func encodeU64(n uint64) []byte {
	result := make([]byte, 8)
	for i := 0; i < 8; i++ {
		result[i] = byte(n >> (8 * i))
	}
	return result
}
