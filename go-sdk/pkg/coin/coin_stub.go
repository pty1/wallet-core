//go:build !cgo
// +build !cgo

// Stub implementation for testing without CGO

package coin

import "fmt"

// GetName returns the human-readable name
func (c CoinType) GetName() string {
	names := map[CoinType]string{
		0:   "Bitcoin",
		2:   "Litecoin",
		3:   "Dogecoin",
		60:  "Ethereum",
		195: "Tron",
	}
	if name, ok := names[c]; ok {
		return name
	}
	return fmt.Sprintf("Coin(%d)", c)
}

// Decimals returns the number of decimal places
func (c CoinType) Decimals() int {
	decimals := map[CoinType]int{
		0:   8,
		2:   8,
		3:   8,
		60:  18,
		195: 6,
	}
	if d, ok := decimals[c]; ok {
		return d
	}
	return 18
}

// Symbol returns the symbol/ticker
func (c CoinType) Symbol() string {
	symbols := map[CoinType]string{
		0:   "BTC",
		2:   "LTC",
		3:   "DOGE",
		60:  "ETH",
		195: "TRX",
	}
	if s, ok := symbols[c]; ok {
		return s
	}
	return "?"
}

// DerivationPath returns the default derivation path
func (c CoinType) DerivationPath() string {
	return fmt.Sprintf("m/44'/%d'/0'/0/0", c)
}

// String returns a string representation
func (c CoinType) String() string {
	return fmt.Sprintf("%s (%s)", c.GetName(), c.Symbol())
}

// TWStringGoString is a no-op stub
func TWStringGoString(s interface{}) string {
	return ""
}
