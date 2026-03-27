//go:build cgo
// +build cgo

package wallet

/*
#include <TrustWalletCore/TWHDWallet.h>
#include <TrustWalletCore/TWPrivateKey.h>
#include <TrustWalletCore/TWPublicKey.h>
#include <TrustWalletCore/TWData.h>
#include <TrustWalletCore/TWString.h>
*/
import "C"
import (
	"encoding/hex"
	"errors"
	"unsafe"

	"github.com/trustwallet/go-wallet-core/pkg/coin"
)

type Wallet struct {
	ptr *C.struct_TWHDWallet
}

func NewWalletFromMnemonic(mnemonic string) (*Wallet, error) {
	if mnemonic == "" {
		return nil, errors.New("mnemonic cannot be empty")
	}

	cMnemonic := C.TWStringCreateWithUTF8Bytes(C.CString(mnemonic))
	defer C.TWStringDelete(cMnemonic)

	cPassword := C.TWStringCreateWithUTF8Bytes(C.CString(""))
	defer C.TWStringDelete(cPassword)

	walletPtr := C.TWHDWalletCreateWithMnemonic(cMnemonic, cPassword)
	if walletPtr == nil {
		return nil, errors.New("failed to create wallet from mnemonic")
	}

	return &Wallet{ptr: walletPtr}, nil
}

func (w *Wallet) Derive(ct coin.CoinType) (*Account, error) {
	if w.ptr == nil {
		return nil, errors.New("wallet not initialized")
	}

	privateKey := C.TWHDWalletGetKeyForCoin(w.ptr, C.enum_TWCoinType(ct))
	defer C.TWPrivateKeyDelete(privateKey)

	address := C.TWHDWalletGetAddressForCoin(w.ptr, C.enum_TWCoinType(ct))
	defer C.TWStringDelete(address)

	pubKey := C.TWPrivateKeyGetPublicKeySecp256k1(privateKey, true)
	defer C.TWPublicKeyDelete(pubKey)
	pubKeyData := C.TWPublicKeyData(pubKey)
	defer C.TWDataDelete(pubKeyData)

	priKeyData := C.TWPrivateKeyData(privateKey)
	defer C.TWDataDelete(priKeyData)

	return &Account{
		coinType: ct,
		address:  TWStringGoString(address),
		pubKey:   hex.EncodeToString(TWDataGoBytes(pubKeyData)),
		priKey:   hex.EncodeToString(TWDataGoBytes(priKeyData)),
		wallet:   w,
	}, nil
}

func (w *Wallet) Close() {
	if w.ptr != nil {
		C.TWHDWalletDelete(w.ptr)
		w.ptr = nil
	}
}

type Account struct {
	coinType coin.CoinType
	address  string
	pubKey   string
	priKey   string
	wallet   *Wallet
}

func (a *Account) Address() string {
	return a.address
}

func (a *Account) PublicKey() string {
	return a.pubKey
}

func (a *Account) PrivateKey() string {
	return a.priKey
}

func (a *Account) CoinType() coin.CoinType {
	return a.coinType
}

func (a *Account) SignTransaction(data []byte) ([]byte, error) {
	return nil, errors.New("use transaction package for signing")
}

func TWStringGoString(s unsafe.Pointer) string {
	if s == nil {
		return ""
	}
	data := C.TWStringUTF8Bytes(s)
	if data == nil {
		return ""
	}
	return C.GoString(data)
}

func TWDataGoBytes(d unsafe.Pointer) []byte {
	cBytes := C.TWDataBytes(d)
	cSize := C.TWDataSize(d)
	return C.GoBytes(unsafe.Pointer(cBytes), C.int(cSize))
}
