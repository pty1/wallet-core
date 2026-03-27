//go:build cgo
// +build cgo

// Package coin provides coin type definitions and utilities for the Trust Wallet Core SDK.
package coin

/*
#include <TrustWalletCore/TWCoinType.h>
#include <TrustWalletCore/TWCoinTypeConfiguration.h>
#include <TrustWalletCore/TWString.h>
*/
import "C"
import (
	"fmt"
)

func (c CoinType) GetName() string {
	name := C.TWCoinTypeConfigurationGetName(C.enum_TWCoinType(c))
	defer C.TWStringDelete(name)
	return TWStringGoString(name)
}

// Decimals returns the number of decimal places for the coin.
func (c CoinType) Decimals() int {
	return int(C.TWCoinTypeConfigurationGetDecimals(C.enum_TWCoinType(c)))
}

// Symbol returns the symbol/ticker of the coin (e.g., "BTC", "ETH").
func (c CoinType) Symbol() string {
	symbol := C.TWCoinTypeConfigurationGetSymbol(C.enum_TWCoinType(c))
	defer C.TWStringDelete(symbol)
	return TWStringGoString(symbol)
}

func (c CoinType) ID() string {
	id := C.TWCoinTypeConfigurationGetID(C.enum_TWCoinType(c))
	defer C.TWStringDelete(id)
	return TWStringGoString(id)
}

func (c CoinType) String() string {
	return fmt.Sprintf("%s (%s)", c.GetName(), c.Symbol())
}
