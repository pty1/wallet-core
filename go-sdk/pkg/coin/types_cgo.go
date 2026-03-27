//go:build cgo
// +build cgo

// Code generated from registry.json. DO NOT EDIT.

package coin

// #include <TrustWalletCore/TWString.h>
import "C"
import "unsafe"

// TWStringGoString converts a TWString to a Go string.
// This is an internal helper function.
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
