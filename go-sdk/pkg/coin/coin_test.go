package coin

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCoinTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		coinType CoinType
		wantID   int
	}{
		{"Bitcoin", Bitcoin, 0},
		{"Litecoin", Litecoin, 2},
		{"Dogecoin", Doge, 3},
		{"Ethereum", Ethereum, 60},
		{"Tron", Tron, 195},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, CoinType(tt.wantID), tt.coinType)
		})
	}
}

func TestCoinByID(t *testing.T) {
	tests := []struct {
		id      string
		want    CoinType
		wantOK  bool
	}{
		{"bitcoin", Bitcoin, true},
		{"ethereum", Ethereum, true},
		{"invalid", 0, false},
		{"", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got, ok := CoinByID(tt.id)
			assert.Equal(t, tt.wantOK, ok)
			if tt.wantOK {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestCoinBySymbol(t *testing.T) {
	tests := []struct {
		symbol string
		wantOK bool
	}{
		{"BTC", true},
		{"ETH", true},
		{"DOGE", true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.symbol, func(t *testing.T) {
			_, ok := CoinBySymbol(tt.symbol)
			assert.Equal(t, tt.wantOK, ok)
		})
	}
}

func TestAllCoins(t *testing.T) {
	coins := AllCoins()
	assert.NotEmpty(t, coins)
	assert.GreaterOrEqual(t, len(coins), 4)

	// Verify some known coins are included
	foundBTC, foundETH := false, false
	for _, c := range coins {
		if c == Bitcoin {
			foundBTC = true
		}
		if c == Ethereum {
			foundETH = true
		}
	}
	assert.True(t, foundBTC, "Bitcoin should be in AllCoins")
	assert.True(t, foundETH, "Ethereum should be in AllCoins")
}
