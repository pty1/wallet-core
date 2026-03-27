package core


// #include <TrustWalletCore/TWMnemonic.h>
import "C"

import "tw/types"

func IsMnemonicValid(mn string) bool {
	str := types.TWStringCreateWithGoString(mn)
	defer C.TWStringDelete(str)
	return bool(C.TWMnemonicIsValid(str))
}
