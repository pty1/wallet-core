package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"math/big"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	"github.com/trustwallet/go-wallet-core/pkg/transaction"
	"github.com/trustwallet/go-wallet-core/pkg/wallet"
)

func main() {
	fmt.Println("Trust Wallet Core Go SDK - Multi-Chain Transactions")
	fmt.Println("====================================================")

	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	w, err := wallet.NewWalletFromMnemonic(mnemonic)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}
	defer w.Close()

	coins := []coin.CoinType{
		coin.Bitcoin,
		coin.Ethereum,
		coin.Litecoin,
		coin.Tron,
		coin.Doge,
		coin.Solana,
		coin.Sui,
	}

	for _, ct := range coins {
		fmt.Printf("\n--- %s (%s) ---\n", ct.GetName(), ct.Symbol())

		account, err := w.Derive(ct)
		if err != nil {
			fmt.Printf("  Error deriving account: %v\n", err)
			continue
		}

		fmt.Printf("  Address: %s\n", account.Address())
		fmt.Printf("  Public Key: %s\n", account.PublicKey()[:32]+"...")

		// Try to build a transaction for this coin
		switch ct {
		case coin.Bitcoin, coin.Litecoin, coin.Doge:
			buildBitcoinLikeTx(ct, account)
		case coin.Ethereum:
			buildEthereumTx(account)
		case coin.Tron:
			fmt.Println("  Tron transaction: (requires specific proto setup)")
		case coin.Solana:
			fmt.Println("  Solana transaction: (requires specific proto setup)")
		case coin.Sui:
			fmt.Println("  Sui transaction: (requires specific proto setup)")
		}
	}
}

func buildBitcoinLikeTx(ct coin.CoinType, account *wallet.Account) {
	fmt.Println("\n  Building Bitcoin-like transaction...")

	// Example UTXO (would normally come from blockchain)
	utxoHash, _ := hex.DecodeString("fff7f7881a8099afa6940d42d1e7f6362bec38171ea3edf433541db4e4ad969f")

	txBuilder := transaction.NewBitcoinTransaction().
		CoinType(ct).
		To(account.Address()).
		Change(account.Address()).
		Amount(100000).
		FeeRate(10).
		PrivateKeys([][]byte{parseHex(account.PrivateKey())}).
		AddUTXO(transaction.BitcoinUTXO{
			TxHash:   utxoHash,
			TxIndex:  0,
			Amount:   200000,
			Script:   []byte{}, // Would need to get lock script
			Sequence: 4294967295,
		})

	// Note: This will fail without proper UTXO data, but shows the API
	_, err := txBuilder.Sign()
	if err != nil {
		fmt.Printf("  Transaction signing requires valid UTXO: %v\n", err)
	}
}

func buildEthereumTx(account *wallet.Account) {
	fmt.Println("\n  Building Ethereum transaction...")

	chainID := big.NewInt(1) // Ethereum mainnet
	value := big.NewInt(1000000000000000000) // 1 ETH in wei
	gasPrice := big.NewInt(1000000000) // 1 Gwei
	gasLimit := uint64(21000)

	txBuilder := transaction.NewEthereumTransaction().
		ChainID(chainID).
		Nonce(0).
		GasLimit(gasLimit).
		To("0x1234567890123456789012345678901234567890").
		Value(value).
		GasPrice(gasPrice)

	// Note: This requires CGO and the wallet-core library
	signed, err := txBuilder.Sign(parseHex(account.PrivateKey()))
	if err != nil {
		fmt.Printf("  Transaction signing requires CGO: %v\n", err)
		return
	}

	fmt.Printf("  Signed transaction: %x\n", signed[:min(64, len(signed))])
}

func parseHex(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		return []byte{}
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
