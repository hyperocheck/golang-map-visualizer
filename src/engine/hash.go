package engine

import (
	"unsafe"
)

//go:noescape
func GetMapType(m any) uintptr

type hashFunc func(unsafe.Pointer, uintptr) uintptr

var mapType *maptype

func CheckHash[K comparable, V any](t *Meta[K, V], key K) uint8 {
	if mapType == nil {
		addr := GetMapType(t.Map)
		mapType = (*maptype)(unsafe.Pointer(addr))
	}

	hmap := GetHmap(t.Map)

	var hf hashFunc
	*(*uintptr)(unsafe.Pointer(&hf)) = *(*uintptr)(unsafe.Pointer(&mapType.Hasher))

	keyPtr := unsafe.Pointer(&key)
	seed := uintptr(hmap.Hash0)
	hash := hf(keyPtr, seed)

	return lowerbytes(hash, hmap.B)
}

func lowerbytes(hash uintptr, B uint8) uint8 {
	return uint8(hash & ((uintptr(1) << B) - 1))
}

func tophash(hash uintptr) uint8 {
	top := uint8(hash >> (8 * (unsafe.Sizeof(uintptr(0)) - 1)))
	if top < 5 {
		top += 5
	}
	return top
}
