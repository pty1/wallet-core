package main

import (
	"fmt"
	"log"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	"github.com/trustwallet/go-wallet-core/pkg/wallet"
)

func main() {
	fmt.Println("Trust Wallet Core Go SDK - Basic Example")
	fmt.Println("=========================================")

	// Demonstrate coin types
	fmt.Println("\nSupported Coins:")
	fmt.Printf("- Bitcoin: ID=%d, Symbol=%s\n", coin.Bitcoin, coin.Bitcoin.Symbol())
	fmt.Printf("- Ethereum: ID=%d, Symbol=%s\n", coin.Ethereum, coin.Ethereum.Symbol())
	fmt.Printf("- Dogecoin: ID=%d, Symbol=%s\n", coin.Doge, coin.Doge.Symbol())

	// Lookup coins
	if btc, ok := coin.CoinByID("bitcoin"); ok {
		fmt.Printf("\nFound coin by ID 'bitcoin': %d\n", btc)
	}

	if eth, ok := coin.CoinBySymbol("ETH"); ok {
		fmt.Printf("Found coin by symbol 'ETH': %d\n", eth)
	}

	// Show all coins
	allCoins := coin.AllCoins()
	fmt.Printf("\nTotal coins supported: %d\n", len(allCoins))

	// Create wallet
	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
	w, err := wallet.NewWalletFromMnemonic(mnemonic)
	if err != nil {
		log.Fatal(err)
	}

	// Derive accounts
	fmt.Println("\nDerived Accounts:")
	
	btcAccount, err := w.Derive(coin.Bitcoin)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("- BTC: %s (Coin: %s)\n", btcAccount.Address(), btcAccount.CoinType().Symbol())

	ethAccount, err := w.Derive(coin.Ethereum)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("- ETH: %s (Coin: %s)\n", ethAccount.Address(), ethAccount.CoinType().Symbol())

	fmt.Println("\nExample completed successfully!")
}
