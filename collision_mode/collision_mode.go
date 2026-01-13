package collision_mode

import (
	"fmt"
	"unsafe"

	"visualizer/src/engine"
	"visualizer/src/logger"
)

var (
    mapType *_maptype_
)

func CheckHash[K comparable, V any](t *engine.Type[K, V], key K) uint8 {
	
	if mapType == nil {
		addr := GetMapType(t.Data)
		mapType = (*_maptype_)(unsafe.Pointer(addr))
	}

	hmap := t.GetHmap()

	var hf hashFunc
	*(*uintptr)(unsafe.Pointer(&hf)) = *(*uintptr)(unsafe.Pointer(&mapType.Hasher))
	
	keyPtr := unsafe.Pointer(&key)
	seed := uintptr(hmap.Hash0) 
	hash := hf(keyPtr, seed)

	//top := tophash(hash) 
	logger.Log.Log("INFO", fmt.Sprintf("key: %d, tophash: %d", key, lowerbytes(hash, hmap.B)))

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
