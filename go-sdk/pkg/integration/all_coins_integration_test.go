//go:build cgo
// +build cgo

package integration

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/trustwallet/go-wallet-core/pkg/coin"
	"github.com/trustwallet/go-wallet-core/pkg/transaction"
	"github.com/trustwallet/go-wallet-core/pkg/wallet"
)

// Test mnemonic for all integration tests
const testMnemonic = "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

// ChainTestCase represents a test case for a specific chain
type ChainTestCase struct {
	Name            string
	CoinType        coin.CoinType
	ExpectedAddress string
	SignTransaction func(*wallet.Account) ([]byte, error)
}

// requireValidSignedTx verifies that a signed transaction is valid
func requireValidSignedTx(t *testing.T, name string, signedTx []byte) {
	require.NotEmpty(t, signedTx, "%s: signed transaction must not be empty", name)
	require.Greater(t, len(signedTx), 10, "%s: signed transaction too short (got %d bytes)", name, len(signedTx))
}

// AllCoinsTestCases contains test cases for all 164 supported coins
// Every coin MUST support transaction signing - failures are not tolerated
var AllCoinsTestCases = []ChainTestCase{
	// Bitcoin Family
	{Name: "Bitcoin", CoinType: coin.Bitcoin, SignTransaction: signBitcoinTx},
	{Name: "Litecoin", CoinType: coin.Litecoin, SignTransaction: signLitecoinTx},
	{Name: "Dogecoin", CoinType: coin.Doge, SignTransaction: signDogecoinTx},
	{Name: "Dash", CoinType: coin.Dash, SignTransaction: signDashTx},
	{Name: "Viacoin", CoinType: coin.Viacoin, SignTransaction: signViacoinTx},
	{Name: "Groestlcoin", CoinType: coin.Groestlcoin, SignTransaction: signGroestlcoinTx},
	{Name: "DigiByte", CoinType: coin.Digibyte, SignTransaction: signDigiByteTx},
	{Name: "Monacoin", CoinType: coin.Monacoin, SignTransaction: signMonacoinTx},
	{Name: "Decred", CoinType: coin.Decred, SignTransaction: signDecredTx},
	{Name: "Syscoin", CoinType: coin.Syscoin, SignTransaction: signSyscoinTx},
	{Name: "Firo", CoinType: coin.Firo, SignTransaction: signFiroTx},
	{Name: "Pivx", CoinType: coin.Pivx, SignTransaction: signPivxTx},
	{Name: "Qtum", CoinType: coin.Qtum, SignTransaction: signQtumTx},
	{Name: "Ravencoin", CoinType: coin.Ravencoin, SignTransaction: signRavencoinTx},
	{Name: "BitcoinGold", CoinType: coin.Bitcoingold, SignTransaction: signBitcoinGoldTx},

	// BitcoinCash Family
	{Name: "BitcoinCash", CoinType: coin.Bitcoincash, SignTransaction: signBitcoinCashTx},
	{Name: "ECash", CoinType: coin.Ecash, SignTransaction: signECashTx},

	// Bitcoin Diamond
	{Name: "BitcoinDiamond", CoinType: coin.Bitcoindiamond, SignTransaction: signBitcoinDiamondTx},

	// Zcash Family
	{Name: "Zcash", CoinType: coin.Zcash, SignTransaction: signZcashTx},
	{Name: "Komodo", CoinType: coin.Komodo, SignTransaction: signKomodoTx},

	// Verge
	{Name: "Verge", CoinType: coin.Verge, SignTransaction: signVergeTx},

	// Ethereum Family
	{Name: "Ethereum", CoinType: coin.Ethereum, SignTransaction: signEthereumTx},
	{Name: "EthereumClassic", CoinType: coin.Classic, SignTransaction: signEthereumClassicTx},
	{Name: "Base", CoinType: coin.Base, SignTransaction: signBaseTx},
	{Name: "Linea", CoinType: coin.Linea, SignTransaction: signLineaTx},
	{Name: "Mantle", CoinType: coin.Mantle, SignTransaction: signMantleTx},
	{Name: "ZenEON", CoinType: coin.Zeneon, SignTransaction: signZenEONTx},
	{Name: "AvalancheC", CoinType: coin.Avalanchec, SignTransaction: signAvalancheCTx},
	{Name: "Polygon", CoinType: coin.Polygon, SignTransaction: signPolygonTx},
	{Name: "PolygonzkEVM", CoinType: coin.Polygonzkevm, SignTransaction: signPolygonzkEVMTx},
	{Name: "BSC", CoinType: coin.Bsc, SignTransaction: signBSCTx},
	{Name: "SmartChain", CoinType: coin.Smartchain, SignTransaction: signSmartChainTx},
	{Name: "Boba", CoinType: coin.Boba, SignTransaction: signBobaTx},
	{Name: "Arbitrum", CoinType: coin.Arbitrum, SignTransaction: signArbitrumTx},
	{Name: "ArbitrumNova", CoinType: coin.Arbitrumnova, SignTransaction: signArbitrumNovaTx},
	{Name: "Optimism", CoinType: coin.Optimism, SignTransaction: signOptimismTx},
	{Name: "Fantom", CoinType: coin.Fantom, SignTransaction: signFantomTx},
	{Name: "Cronos", CoinType: coin.Cronos, SignTransaction: signCronosTx},
	{Name: "Celo", CoinType: coin.Celo, SignTransaction: signCeloTx},
	{Name: "Gnosis", CoinType: coin.Xdai, SignTransaction: signGnosisTx},
	{Name: "Rootstock", CoinType: coin.Rootstock, SignTransaction: signRootstockTx},
	{Name: "Wanchain", CoinType: coin.Wanchain, SignTransaction: signWanchainTx},
	{Name: "GoChain", CoinType: coin.Gochain, SignTransaction: signGoChainTx},
	{Name: "KCC", CoinType: coin.Kcc, SignTransaction: signKCCTx},
	{Name: "Moonriver", CoinType: coin.Moonriver, SignTransaction: signMoonriverTx},
	{Name: "Moonbeam", CoinType: coin.Moonbeam, SignTransaction: signMoonbeamTx},
	{Name: "Meter", CoinType: coin.Meter, SignTransaction: signMeterTx},
	{Name: "OKC", CoinType: coin.Okc, SignTransaction: signOKCTx},
	{Name: "ConfluxESpace", CoinType: coin.Cfxevm, SignTransaction: signConfluxESpaceTx},
	{Name: "AcalaEVM", CoinType: coin.Acalaevm, SignTransaction: signAcalaEVMTx},
	{Name: "IoTeXEVM", CoinType: coin.Iotexevm, SignTransaction: signIoTeXEVMTx},
	{Name: "SmartBitcoinCash", CoinType: coin.Smartbch, SignTransaction: signSmartBitcoinCashTx},
	{Name: "ThunderCore", CoinType: coin.Thundertoken, SignTransaction: signThunderCoreTx},
	{Name: "ThetaFuelEVM", CoinType: coin.Tfuelevm, SignTransaction: signThetaFuelEVMTx},
	{Name: "OasisEmerald", CoinType: coin.Oasis, SignTransaction: signOasisEmeraldTx},
	{Name: "Harmony", CoinType: coin.Harmony, SignTransaction: signHarmonyTx},
	{Name: "OPBNB", CoinType: coin.Opbnb, SignTransaction: signOPBNBTx},
	{Name: "ZkSync", CoinType: coin.Zksync, SignTransaction: signZkSyncTx},
	{Name: "Scroll", CoinType: coin.Scroll, SignTransaction: signScrollTx},
	{Name: "Manta", CoinType: coin.Manta, SignTransaction: signMantaTx},
	{Name: "Merlin", CoinType: coin.Merlin, SignTransaction: signMerlinTx},
	{Name: "Blast", CoinType: coin.Blast, SignTransaction: signBlastTx},
	{Name: "ZkLinkNova", CoinType: coin.Zklinknova, SignTransaction: signZkLinkNovaTx},
	{Name: "LightLink", CoinType: coin.Lightlink, SignTransaction: signLightLinkTx},
	{Name: "Metis", CoinType: coin.Metis, SignTransaction: signMetisTx},
	{Name: "Aurora", CoinType: coin.Aurora, SignTransaction: signAuroraTx},
	{Name: "Evmos", CoinType: coin.Evmos, SignTransaction: signEvmosTx},
	{Name: "KavaEVM", CoinType: coin.Kavaevm, SignTransaction: signKavaEVMTx},
	{Name: "POANetwork", CoinType: coin.Poa, SignTransaction: signPOANetworkTx},
	{Name: "Theta", CoinType: coin.Theta, SignTransaction: signThetaTx},
	{Name: "Callisto", CoinType: coin.Callisto, SignTransaction: signCallistoTx},
	{Name: "Ronin", CoinType: coin.Ronin, SignTransaction: signRoninTx},
	{Name: "Viction", CoinType: coin.Viction, SignTransaction: signVictionTx},
	{Name: "Kaia", CoinType: coin.Kaia, SignTransaction: signKaiaTx},
	{Name: "ZetaEVM", CoinType: coin.Zetaevm, SignTransaction: signZetaEVMTx},
	{Name: "MegaETH", CoinType: coin.Megaeth, SignTransaction: signMegaETHTx},
	{Name: "Neon", CoinType: coin.Neon, SignTransaction: signNeonTx},
	{Name: "Heco", CoinType: coin.Heco, SignTransaction: signHecoTx},

	// Cosmos Family
	{Name: "Cosmos", CoinType: coin.Cosmos, SignTransaction: signCosmosTx},
	{Name: "Stargaze", CoinType: coin.Stargaze, SignTransaction: signStargazeTx},
	{Name: "Juno", CoinType: coin.Juno, SignTransaction: signJunoTx},
	{Name: "Stride", CoinType: coin.Stride, SignTransaction: signStrideTx},
	{Name: "Axelar", CoinType: coin.Axelar, SignTransaction: signAxelarTx},
	{Name: "Crescent", CoinType: coin.Crescent, SignTransaction: signCrescentTx},
	{Name: "Kujira", CoinType: coin.Kujira, SignTransaction: signKujiraTx},
	{Name: "Comdex", CoinType: coin.Comdex, SignTransaction: signComdexTx},
	{Name: "Neutron", CoinType: coin.Neutron, SignTransaction: signNeutronTx},
	{Name: "Sommelier", CoinType: coin.Sommelier, SignTransaction: signSommelierTx},
	{Name: "FetchAI", CoinType: coin.Fetchai, SignTransaction: signFetchAITx},
	{Name: "Mars", CoinType: coin.Mars, SignTransaction: signMarsTx},
	{Name: "Umee", CoinType: coin.Umee, SignTransaction: signUmeeTx},
	{Name: "Noble", CoinType: coin.Noble, SignTransaction: signNobleTx},
	{Name: "Sei", CoinType: coin.Sei, SignTransaction: signSeiTx},
	{Name: "Tia", CoinType: coin.Tia, SignTransaction: signTiaTx},
	{Name: "Coreum", CoinType: coin.Coreum, SignTransaction: signCoreumTx},
	{Name: "Quasar", CoinType: coin.Quasar, SignTransaction: signQuasarTx},
	{Name: "Persistence", CoinType: coin.Persistence, SignTransaction: signPersistenceTx},
	{Name: "Akash", CoinType: coin.Akash, SignTransaction: signAkashTx},
	{Name: "Osmosis", CoinType: coin.Osmosis, SignTransaction: signOsmosisTx},
	{Name: "Kava", CoinType: coin.Kava, SignTransaction: signKavaTx},
	{Name: "Band", CoinType: coin.Band, SignTransaction: signBandTx},
	{Name: "Bluzelle", CoinType: coin.Bluzelle, SignTransaction: signBluzelleTx},
	{Name: "CryptoOrg", CoinType: coin.Cryptoorg, SignTransaction: signCryptoOrgTx},
	{Name: "Secret", CoinType: coin.Secret, SignTransaction: signSecretTx},
	{Name: "Terra", CoinType: coin.Terra, SignTransaction: signTerraTx},
	{Name: "TerraV2", CoinType: coin.Terrav2, SignTransaction: signTerraV2Tx},
	{Name: "Agoric", CoinType: coin.Agoric, SignTransaction: signAgoricTx},
	{Name: "DYDX", CoinType: coin.Dydx, SignTransaction: signDYDXTx},
	{Name: "NativeInjective", CoinType: coin.Nativeinjective, SignTransaction: signNativeInjectiveTx},
	{Name: "NativeCanto", CoinType: coin.Nativecanto, SignTransaction: signNativeCantoTx},
	{Name: "NativeEvmos", CoinType: coin.Nativeevmos, SignTransaction: signNativeEvmosTx},
	{Name: "Acala", CoinType: coin.Acala, SignTransaction: signAcalaTx},
	{Name: "THORChain", CoinType: coin.Thorchain, SignTransaction: signTHORChainTx},
	{Name: "ZetaChain", CoinType: coin.Zetachain, SignTransaction: signZetaChainTx},

	// Solana
	{Name: "Solana", CoinType: coin.Solana, SignTransaction: signSolanaTx},

	// Cardano
	{Name: "Cardano", CoinType: coin.Cardano, SignTransaction: signCardanoTx},

	// Polkadot Family
	{Name: "Polkadot", CoinType: coin.Polkadot, SignTransaction: signPolkadotTx},
	{Name: "Kusama", CoinType: coin.Kusama, SignTransaction: signKusamaTx},

	// Ripple
	{Name: "XRP", CoinType: coin.Ripple, SignTransaction: signXRPTx},

	// Stellar Family
	{Name: "Stellar", CoinType: coin.Stellar, SignTransaction: signStellarTx},
	{Name: "Kin", CoinType: coin.Kin, SignTransaction: signKinTx},

	// Tezos
	{Name: "Tezos", CoinType: coin.Tezos, SignTransaction: signTezosTx},

	// Tron
	{Name: "Tron", CoinType: coin.Tron, SignTransaction: signTronTx},

	// EOS Family
	{Name: "EOS", CoinType: coin.Eos, SignTransaction: signEOSTx},
	{Name: "WAX", CoinType: coin.Wax, SignTransaction: signWAXTx},

	// Zelcash
	{Name: "Zelcash", CoinType: coin.Zelcash, SignTransaction: signZelcashTx},

	// Native chains
	{Name: "Aeternity", CoinType: coin.Aeternity, SignTransaction: signAeternityTx},
	{Name: "Aion", CoinType: coin.Aion, SignTransaction: signAionTx},
	{Name: "Algorand", CoinType: coin.Algorand, SignTransaction: signAlgorandTx},
	{Name: "Aptos", CoinType: coin.Aptos, SignTransaction: signAptosTx},
	{Name: "Sui", CoinType: coin.Sui, SignTransaction: signSuiTx},
	{Name: "NEAR", CoinType: coin.Near, SignTransaction: signNEARTx},
	{Name: "Filecoin", CoinType: coin.Filecoin, SignTransaction: signFilecoinTx},
	{Name: "Hedera", CoinType: coin.Hedera, SignTransaction: signHederaTx},
	{Name: "ICON", CoinType: coin.Icon, SignTransaction: signICONTx},
	{Name: "InternetComputer", CoinType: coin.Internet_computer, SignTransaction: signInternetComputerTx},
	{Name: "IOST", CoinType: coin.Iost, SignTransaction: signIOSTTx},
	{Name: "IoTeX", CoinType: coin.Iotex, SignTransaction: signIoTeXTx},
	{Name: "Nano", CoinType: coin.Nano, SignTransaction: signNanoTx},
	{Name: "Nebulas", CoinType: coin.Nebulas, SignTransaction: signNebulasTx},
	{Name: "NEO", CoinType: coin.Neo, SignTransaction: signNEOTx},
	{Name: "Nervos", CoinType: coin.Nervos, SignTransaction: signNervosTx},
	{Name: "Nimiq", CoinType: coin.Nimiq, SignTransaction: signNimiqTx},
	{Name: "Ontology", CoinType: coin.Ontology, SignTransaction: signOntologyTx},
	{Name: "MultiversX", CoinType: coin.Elrond, SignTransaction: signMultiversXTx},
	{Name: "TON", CoinType: coin.Ton, SignTransaction: signTONTx},
	{Name: "VeChain", CoinType: coin.Vechain, SignTransaction: signVeChainTx},
	{Name: "Waves", CoinType: coin.Waves, SignTransaction: signWavesTx},
	{Name: "Zilliqa", CoinType: coin.Zilliqa, SignTransaction: signZilliqaTx},
	{Name: "Zen", CoinType: coin.Zen, SignTransaction: signZenTx},
	{Name: "FIO", CoinType: coin.Fio, SignTransaction: signFIOTx},
	{Name: "Greenfield", CoinType: coin.Greenfield, SignTransaction: signGreenfieldTx},
	{Name: "Everscale", CoinType: coin.Everscale, SignTransaction: signEverscaleTx},
	{Name: "Pactus", CoinType: coin.Pactus, SignTransaction: signPactusTx},
	{Name: "Polymesh", CoinType: coin.Polymesh, SignTransaction: signPolymeshTx},
	{Name: "BounceBit", CoinType: coin.Bouncebit, SignTransaction: signBounceBitTx},
	{Name: "Sonic", CoinType: coin.Sonic, SignTransaction: signSonicTx},
	{Name: "Stratis", CoinType: coin.Stratis, SignTransaction: signStratisTx},
	{Name: "Neblio", CoinType: coin.Nebl, SignTransaction: signNeblioTx},
	{Name: "Plasma", CoinType: coin.Plasma, SignTransaction: signPlasmaTx},
	{Name: "Monad", CoinType: coin.Monad, SignTransaction: signMonadTx},
	{Name: "BinanceChain", CoinType: coin.Binance, SignTransaction: signBinanceChainTx},
	{Name: "TestBinance", CoinType: coin.Tbinance, SignTransaction: signTestBinanceTx},
	{Name: "NULS", CoinType: coin.Nuls, SignTransaction: signNULSTx},
}

// TestAllCoins_AddressDerivation tests address derivation for all 164 coins
func TestAllCoins_AddressDerivation(t *testing.T) {
	w, err := wallet.NewWalletFromMnemonic(testMnemonic)
	require.NoError(t, err)
	defer w.Close()

	for _, tc := range AllCoinsTestCases {
		t.Run(tc.Name, func(t *testing.T) {
			account, err := w.Derive(tc.CoinType)
			require.NoError(t, err, "Failed to derive address for %s", tc.Name)

			assert.NotEmpty(t, account.Address(), "Address should not be empty for %s", tc.Name)
			assert.Equal(t, tc.CoinType, account.CoinType(), "Coin type mismatch for %s", tc.Name)
			assert.NotEmpty(t, account.PublicKey(), "Public key should not be empty for %s", tc.Name)
			assert.NotEmpty(t, account.PrivateKey(), "Private key should not be empty for %s", tc.Name)

			if tc.ExpectedAddress != "" {
				assert.Equal(t, tc.ExpectedAddress, account.Address(), "Address mismatch for %s", tc.Name)
			}

			t.Logf("✓ %s: %s (%s)", tc.Name, account.Address(), tc.CoinType.Symbol())
		})
	}
}

// TestAllCoins_TransactionSigning tests transaction signing for ALL 164 coins
// Every coin MUST support transaction signing - no exceptions, no skips
func TestAllCoins_TransactionSigning(t *testing.T) {
	w, err := wallet.NewWalletFromMnemonic(testMnemonic)
	require.NoError(t, err)
	defer w.Close()

	for _, tc := range AllCoinsTestCases {
		t.Run(fmt.Sprintf("%s_SignTx", tc.Name), func(t *testing.T) {
			account, err := w.Derive(tc.CoinType)
			require.NoError(t, err, "Failed to derive account for %s", tc.Name)

			signedTx, err := tc.SignTransaction(account)
			require.NoError(t, err, "Failed to sign transaction for %s", tc.Name)
			requireValidSignedTx(t, tc.Name, signedTx)

			t.Logf("✓ %s: Transaction signed successfully (%d bytes)", tc.Name, len(signedTx))
		})
	}
}

// Bitcoin Family Transaction Signers
func signBitcoinTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Bitcoin, 100000, 10)
}

func signLitecoinTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Litecoin, 100000, 10)
}

func signDogecoinTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Doge, 100000000, 1)
}

func signDashTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Dash, 100000, 10)
}

func signViacoinTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Viacoin, 100000, 10)
}

func signGroestlcoinTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Groestlcoin, 100000, 10)
}

func signDigiByteTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Digibyte, 100000, 10)
}

func signMonacoinTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Monacoin, 100000, 10)
}

func signDecredTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Decred, 100000, 10)
}

func signSyscoinTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Syscoin, 100000, 10)
}

func signFiroTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Firo, 100000, 10)
}

func signPivxTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Pivx, 100000, 10)
}

func signQtumTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Qtum, 100000, 10)
}

func signRavencoinTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Ravencoin, 100000, 10)
}

func signBitcoinGoldTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Bitcoingold, 100000, 10)
}

func signBitcoinCashTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Bitcoincash, 100000, 10)
}

func signECashTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Ecash, 100000, 10)
}

func signBitcoinDiamondTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Bitcoindiamond, 100000, 10)
}

func signZcashTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Zcash, 100000, 10)
}

func signKomodoTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Komodo, 100000, 10)
}

func signVergeTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Verge, 100000, 10)
}

func signZelcashTx(account *wallet.Account) ([]byte, error) {
	return SignP2PKHTransaction(account, coin.Zelcash, 100000, 10)
}

// Ethereum Family Transaction Signers
func signEthereumTx(account *wallet.Account) ([]byte, error) {
	privateKey, err := hex.DecodeString(account.PrivateKey())
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	tx, err := transaction.NewEthereumTransaction().
		ChainID(big.NewInt(1)).
		Nonce(0).
		GasLimit(21000).
		To("0x1234567890123456789012345678901234567890").
		Value(big.NewInt(1000000000000000000)).
		GasPrice(big.NewInt(1000000000)).
		Sign(privateKey)

	if err != nil {
		return nil, fmt.Errorf("ethereum transaction signing failed: %w", err)
	}
	return tx, nil
}

func signEthereumClassicTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 61)
}

func signBaseTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 8453)
}

func signLineaTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 59144)
}

func signMantleTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 5000)
}

func signZenEONTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 7332)
}

func signAvalancheCTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 43114)
}

func signPolygonTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 137)
}

func signPolygonzkEVMTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 1101)
}

func signBSCTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 56)
}

func signSmartChainTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 56)
}

func signBobaTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 288)
}

func signArbitrumTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 42161)
}

func signArbitrumNovaTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 42170)
}

func signOptimismTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 10)
}

func signFantomTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 250)
}

func signCronosTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 25)
}

func signCeloTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 42220)
}

func signGnosisTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 100)
}

func signRootstockTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 30)
}

func signWanchainTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 888)
}

func signGoChainTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 60)
}

func signKCCTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 321)
}

func signMoonriverTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 1285)
}

func signMoonbeamTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 1284)
}

func signMeterTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 82)
}

func signOKCTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 66)
}

func signConfluxESpaceTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 1030)
}

func signAcalaEVMTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 787)
}

func signIoTeXEVMTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 4689)
}

func signSmartBitcoinCashTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 10000)
}

func signThunderCoreTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 108)
}

func signThetaFuelEVMTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 361)
}

func signOasisEmeraldTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 42262)
}

func signHarmonyTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 1666600000)
}

func signOPBNBTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 204)
}

func signZkSyncTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 324)
}

func signScrollTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 534352)
}

func signMantaTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 169)
}

func signMerlinTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 4200)
}

func signBlastTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 81457)
}

func signZkLinkNovaTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 810180)
}

func signLightLinkTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 1890)
}

func signMetisTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 1088)
}

func signAuroraTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 1313161554)
}

func signEvmosTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 9001)
}

func signKavaEVMTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 2222)
}

func signPOANetworkTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 99)
}

func signThetaTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 361)
}

func signCallistoTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 820)
}

func signRoninTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 2020)
}

func signVictionTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 88)
}

func signKaiaTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 8217)
}

func signZetaEVMTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 7000)
}

func signMegaETHTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 4326)
}

func signNeonTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 245022934)
}

func signHecoTx(account *wallet.Account) ([]byte, error) {
	return signEVMTx(account, 128)
}

// signEVMTx is a helper for EVM-compatible chains
func signEVMTx(account *wallet.Account, chainID int64) ([]byte, error) {
	privateKey, err := hex.DecodeString(account.PrivateKey())
	if err != nil {
		return nil, fmt.Errorf("failed to decode private key: %w", err)
	}

	tx, err := transaction.NewEthereumTransaction().
		ChainID(big.NewInt(chainID)).
		Nonce(0).
		GasLimit(21000).
		To("0x1234567890123456789012345678901234567890").
		Value(big.NewInt(1000000000000000000)).
		GasPrice(big.NewInt(1000000000)).
		Sign(privateKey)

	if err != nil {
		return nil, fmt.Errorf("EVM transaction signing failed (chainID: %d): %w", chainID, err)
	}
	return tx, nil
}

// Cosmos Family Transaction Signers
func signCosmosTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Cosmos, "1000000")
}

func signStargazeTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Stargaze, "1000000")
}

func signJunoTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Juno, "1000000")
}

func signStrideTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Stride, "1000000")
}

func signAxelarTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Axelar, "1000000")
}

func signCrescentTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Crescent, "1000000")
}

func signKujiraTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Kujira, "1000000")
}

func signComdexTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Comdex, "1000000")
}

func signNeutronTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Neutron, "1000000")
}

func signSommelierTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Sommelier, "1000000")
}

func signFetchAITx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Fetchai, "1000000")
}

func signMarsTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Mars, "1000000")
}

func signUmeeTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Umee, "1000000")
}

func signNobleTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Noble, "1000000")
}

func signSeiTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Sei, "1000000")
}

func signTiaTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Tia, "1000000")
}

func signCoreumTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Coreum, "1000000")
}

func signQuasarTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Quasar, "1000000")
}

func signPersistenceTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Persistence, "1000000")
}

func signAkashTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Akash, "1000000")
}

func signOsmosisTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Osmosis, "1000000")
}

func signKavaTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Kava, "1000000")
}

func signBandTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Band, "1000000")
}

func signBluzelleTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Bluzelle, "1000000")
}

func signCryptoOrgTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Cryptoorg, "1000000")
}

func signSecretTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Secret, "1000000")
}

func signTerraTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Terra, "1000000")
}

func signTerraV2Tx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Terrav2, "1000000")
}

func signAgoricTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Agoric, "1000000")
}

func signDYDXTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Dydx, "1000000")
}

func signNativeInjectiveTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Nativeinjective, "1000000")
}

func signNativeCantoTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Nativecanto, "1000000")
}

func signNativeEvmosTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Nativeevmos, "1000000")
}

func signAcalaTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Acala, "1000000")
}

func signTHORChainTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Thorchain, "1000000")
}

func signZetaChainTx(account *wallet.Account) ([]byte, error) {
	return SignCosmosTransaction(account, coin.Zetachain, "1000000")
}

// Native Chain Transaction Signers
func signSolanaTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Solana transaction signing not implemented")
}

func signCardanoTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Cardano transaction signing not implemented")
}

func signPolkadotTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Polkadot transaction signing not implemented")
}

func signKusamaTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Kusama transaction signing not implemented")
}

func signXRPTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("XRP transaction signing not implemented")
}

func signStellarTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Stellar transaction signing not implemented")
}

func signKinTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Kin transaction signing not implemented")
}

func signTezosTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Tezos transaction signing not implemented")
}

func signTronTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Tron transaction signing not implemented")
}

func signEOSTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("EOS transaction signing not implemented")
}

func signWAXTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("WAX transaction signing not implemented")
}

func signAeternityTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Aeternity transaction signing not implemented")
}

func signAionTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Aion transaction signing not implemented")
}

func signAlgorandTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Algorand transaction signing not implemented")
}

func signAptosTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Aptos transaction signing not implemented")
}

func signSuiTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Sui transaction signing not implemented")
}

func signNEARTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("NEAR transaction signing not implemented")
}

func signFilecoinTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Filecoin transaction signing not implemented")
}

func signHederaTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Hedera transaction signing not implemented")
}

func signICONTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("ICON transaction signing not implemented")
}

func signInternetComputerTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("InternetComputer transaction signing not implemented")
}

func signIOSTTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("IOST transaction signing not implemented")
}

func signIoTeXTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("IoTeX transaction signing not implemented")
}

func signNanoTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Nano transaction signing not implemented")
}

func signNebulasTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Nebulas transaction signing not implemented")
}

func signNEOTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("NEO transaction signing not implemented")
}

func signNervosTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Nervos transaction signing not implemented")
}

func signNimiqTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Nimiq transaction signing not implemented")
}

func signOntologyTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Ontology transaction signing not implemented")
}

func signMultiversXTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("MultiversX transaction signing not implemented")
}

func signTONTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("TON transaction signing not implemented")
}

func signVeChainTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("VeChain transaction signing not implemented")
}

func signWavesTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Waves transaction signing not implemented")
}

func signZilliqaTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Zilliqa transaction signing not implemented")
}

func signZenTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Zen transaction signing not implemented")
}

func signFIOTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("FIO transaction signing not implemented")
}

func signGreenfieldTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Greenfield transaction signing not implemented")
}

func signEverscaleTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Everscale transaction signing not implemented")
}

func signPactusTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Pactus transaction signing not implemented")
}

func signPolymeshTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Polymesh transaction signing not implemented")
}

func signBounceBitTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("BounceBit transaction signing not implemented")
}

func signSonicTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Sonic transaction signing not implemented")
}

func signStratisTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Stratis transaction signing not implemented")
}

func signNeblioTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Neblio transaction signing not implemented")
}

func signPlasmaTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Plasma transaction signing not implemented")
}

func signMonadTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("Monad transaction signing not implemented")
}

func signBinanceChainTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("BinanceChain transaction signing not implemented")
}

func signTestBinanceTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("TestBinance transaction signing not implemented")
}

func signNULSTx(account *wallet.Account) ([]byte, error) {
	return nil, fmt.Errorf("NULS transaction signing not implemented")
}
