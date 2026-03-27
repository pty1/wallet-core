//go:build cgo
// +build cgo

package integration

import (
	"encoding/hex"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	"github.com/trustwallet/go-wallet-core/pkg/transaction"
	"github.com/trustwallet/go-wallet-core/pkg/wallet"
)

type TestUTXO struct {
	TxHash  string
	TxIndex uint32
	Amount  int64
}

func BuildTestUTXOs(account *wallet.Account, coinType coin.CoinType, utxos []TestUTXO) ([]transaction.BitcoinUTXO, error) {
	lockScript, err := transaction.BuildLockScript(account.Address(), coinType)
	if err != nil {
		return nil, err
	}

	result := make([]transaction.BitcoinUTXO, len(utxos))
	for i, u := range utxos {
		txHash, err := hex.DecodeString(u.TxHash)
		if err != nil {
			return nil, err
		}
		result[i] = transaction.BitcoinUTXO{
			TxHash:  txHash,
			TxIndex: u.TxIndex,
			Amount:  u.Amount,
			Script:  lockScript,
		}
	}
	return result, nil
}

var DefaultTestUTXOs = []TestUTXO{
	{TxHash: "0000000000000000000000000000000000000000000000000000000000000001", TxIndex: 0, Amount: 200000},
}

func SignP2PKHTransaction(account *wallet.Account, coinType coin.CoinType, amount, feeRate int64) ([]byte, error) {
	privateKey, err := hex.DecodeString(account.PrivateKey())
	if err != nil {
		return nil, err
	}

	utxos, err := BuildTestUTXOs(account, coinType, DefaultTestUTXOs)
	if err != nil {
		return nil, err
	}

	result, err := transaction.NewBitcoinTransaction().
		CoinType(coinType).
		To(account.Address()).
		Change(account.Address()).
		Amount(amount).
		FeeRate(feeRate).
		PrivateKeys([][]byte{privateKey}).
		AddUTXO(utxos[0]).
		SignWithResult()

	if err != nil {
		return nil, err
	}
	return result.Encoded, nil
}

func SignCosmosTransaction(account *wallet.Account, coinType coin.CoinType, amount string) ([]byte, error) {
	privateKey, err := hex.DecodeString(account.PrivateKey())
	if err != nil {
		return nil, err
	}

	publicKey, err := transaction.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return nil, err
	}

	return transaction.SignCosmosTransaction(
		coinType,
		privateKey,
		publicKey,
		account.Address(),
		account.Address(),
		amount,
		0,
		0,
	)
}

// Helper for native chains that just need private key
func SignWithPrivateKey(account *wallet.Account, signer func([]byte) ([]byte, error)) ([]byte, error) {
	privateKey, err := hex.DecodeString(account.PrivateKey())
	if err != nil {
		return nil, err
	}
	return signer(privateKey)
}

// Helper for native chains that need private and public key
func SignWithKeyPair(account *wallet.Account, signer func([]byte, []byte) ([]byte, error)) ([]byte, error) {
	privateKey, err := hex.DecodeString(account.PrivateKey())
	if err != nil {
		return nil, err
	}

	publicKey, err := transaction.PrivateKeyToPublicKey(privateKey)
	if err != nil {
		return nil, err
	}

	return signer(privateKey, publicKey)
}
