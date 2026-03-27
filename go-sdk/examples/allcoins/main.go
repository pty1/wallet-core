package main

import (
	"fmt"
	"log"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
	"github.com/trustwallet/go-wallet-core/pkg/wallet"
)

func main() {
	fmt.Println("Trust Wallet Core Go SDK - All 164 Coins")
	fmt.Println("==========================================")

	mnemonic := "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

	w, err := wallet.NewWalletFromMnemonic(mnemonic)
	if err != nil {
		log.Fatalf("Failed to create wallet: %v", err)
	}
	defer w.Close()

	fmt.Printf("\nDeriving addresses for all %d supported coins:\n\n", len(coin.AllCoins()))

	successCount := 0
	failCount := 0

	for i, ct := range coin.AllCoins() {
		account, err := w.Derive(ct)
		if err != nil {
			fmt.Printf("%3d. %-20s ERROR: %v\n", i+1, ct.ID(), err)
			failCount++
			continue
		}

		fmt.Printf("%3d. %-20s %-8s Decimals:%2d Address: %s\n",
			i+1,
			ct.ID(),
			ct.Symbol(),
			ct.Decimals(),
			account.Address(),
		)
		successCount++
	}

	fmt.Printf("\n==========================================\n")
	fmt.Printf("Successfully derived: %d\n", successCount)
	fmt.Printf("Failed: %d\n", failCount)
}
