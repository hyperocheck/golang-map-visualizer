package collision_mode

import (
	"unsafe"
)

//go:noescape
func GetMapType(m interface{}) uintptr

type hashFunc func(unsafe.Pointer, uintptr) uintptr

type _type struct {
	Size_       uintptr
	PtrBytes    uintptr
	Hash        uint32
	TFlag       uint8
	Align_      uint8
	FieldAlign_ uint8
	Kind_       uint8
	Equal       func(unsafe.Pointer, unsafe.Pointer) bool
	GCData      *byte
	Str         int32
	PtrToThis   int32
}

type _maptype_ struct {
	_type
	Key        *_type
	Elem       *_type
	Group      *_type
	Hasher     func(unsafe.Pointer, uintptr) uintptr
	GroupSize  uintptr
	SlotSize   uintptr
	ElemOff    uintptr
	Flags      uint32
}

