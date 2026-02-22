package main

import (
	"fmt"
	"log"
	"unsafe"
)

type Map struct {
	used              uint64
	seed              uintptr
	dirPtr            unsafe.Pointer
	dirLen            int
	globalDepth       uint8
	globalShift       uint8
	writing           uint8
	tombstonePossible bool
	clearSeq          uint64
}

type ctrlGroup uint64

type groupReference struct {
	data unsafe.Pointer
}

type slot[K comparable, V any] struct {
	key  K
	elem V
}

type group[K comparable, V any] struct {
	ctrls ctrlGroup
	slots [8]slot[K, V]
}

type groupsReference struct {
	data unsafe.Pointer

	lengthMask uint64
}

type table struct {
	used       uint16
	capacity   uint16
	growthLeft uint16
	localDepth uint8
	index      int
	groups     groupsReference
}

type probeSeq struct {
	mask   uint64
	offset uint64
	index  uint64
}

func makeProbeSeq(hash uintptr, mask uint64) probeSeq {
	return probeSeq{
		mask:   mask,
		offset: uint64(hash) & mask,
		index:  0,
	}
}

func (s probeSeq) next() probeSeq {
	s.index++
	s.offset = (s.offset + s.index) & s.mask
	return s
}

type bitset uint64

func (b bitset) first() uintptr {
	return bitsetFirst(b)
}

const deBruijn64 = 0x03f79d71b4ca8b09

var deBruijn64tab = [64]byte{
	0, 1, 56, 2, 57, 49, 28, 3, 61, 58, 42, 50, 38, 29, 17, 4,
	62, 47, 59, 36, 45, 43, 51, 22, 53, 39, 33, 30, 24, 18, 12, 5,
	63, 55, 48, 27, 60, 41, 37, 16, 46, 35, 44, 21, 52, 32, 23, 11,
	54, 26, 40, 15, 34, 20, 31, 10, 25, 14, 19, 9, 13, 8, 7, 6,
}

func TrailingZeros64(x uint64) int {
	if x == 0 {
		return 64
	}

	return int(deBruijn64tab[(x&-x)*deBruijn64>>(64-6)])
}

func bitsetFirst(b bitset) uintptr {
	return uintptr(TrailingZeros64(uint64(b))) >> 3
}

type ctrl uint8

const (
	maxAvgGroupLoad = 7

	ctrlEmpty   ctrl = 0b10000000
	ctrlDeleted ctrl = 0b11111110

	bitsetLSB   = 0x0101010101010101
	bitsetMSB   = 0x8080808080808080
	bitsetEmpty = bitsetLSB * uint64(ctrlEmpty)
)

func ctrlGroupMatchH2(g ctrlGroup, h uintptr) bitset {
	v := uint64(g) ^ (bitsetLSB * uint64(h))
	return bitset(((v - bitsetLSB) &^ v) & bitsetMSB)
}

func (g ctrlGroup) matchH2(h uintptr) bitset {
	return ctrlGroupMatchH2(g, h)
}

func (g *groupReference) ctrls() *ctrlGroup {
	return (*ctrlGroup)(g.data)
}

const (
	ctrlGroupsSize   = unsafe.Sizeof(ctrlGroup(0))
	groupSlotsOffset = ctrlGroupsSize
)

func (g *groupReference) key(typ *MapType, i uintptr) unsafe.Pointer {
	offset := groupSlotsOffset + i*typ.SlotSize

	return unsafe.Pointer(uintptr(g.data) + offset)
}

func (g *groupsReference) group(typ *MapType, i uint64) groupReference {
	offset := uintptr(i) * typ.GroupSize

	return groupReference{
		data: unsafe.Pointer(uintptr(g.data) + offset),
	}
}

const (
	MapNeedKeyUpdate = 1 << iota
	MapHashMightPanic
	MapIndirectKey
	MapIndirectElem
)

func (mt *MapType) IndirectKey() bool {
	return mt.Flags&MapIndirectKey != 0
}

func (mt *MapType) IndirectElem() bool {
	return mt.Flags&MapIndirectElem != 0
}

func (g *groupReference) elem(typ *MapType, i uintptr) unsafe.Pointer {
	offset := groupSlotsOffset + i*typ.SlotSize + typ.ElemOff

	return unsafe.Pointer(uintptr(g.data) + offset)
}

func (b bitset) removeFirst() bitset {
	return b & (b - 1)
}

func (g ctrlGroup) matchEmpty() bitset {
	return ctrlGroupMatchEmpty(g)
}

func ctrlGroupMatchEmpty(g ctrlGroup) bitset {
	v := uint64(g)
	return bitset((v &^ (v << 6)) & bitsetMSB)
}

func (t *table) getWithoutKey(typ *MapType, hash uintptr, key unsafe.Pointer) (unsafe.Pointer, bool) {
	seq := makeProbeSeq(h1(hash), t.groups.lengthMask)

	h2Hash := h2(hash)
	for ; ; seq = seq.next() {

		fmt.Printf("\tSEQ ▶️%+v\n", seq)

		g := t.groups.group(typ, seq.offset)

		fmt.Printf("\t\tGROUP ▶️%+v\n", g)

		match := g.ctrls().matchH2(h2Hash)

		for match != 0 {
			i := match.first()

			slotKey := g.key(typ, i)
			if typ.IndirectKey() {
				slotKey = *((*unsafe.Pointer)(slotKey))
			}
			if typ.Key.Equal(key, slotKey) {
				slotElem := g.elem(typ, i)
				if typ.IndirectElem() {
					slotElem = *((*unsafe.Pointer)(slotElem))
				}
				return slotElem, true
			}
			match = match.removeFirst()
		}

		match = g.ctrls().matchEmpty()
		if match != 0 {
			return nil, false
		}
	}
}

func GetMapType(m interface{}) uintptr

type (
	TFlag   uint8
	Kind    uint8
	NameOff int32
	TypeOff int32
)

type Type struct {
	Size_       uintptr
	PtrBytes    uintptr
	Hash        uint32
	TFlag       TFlag
	Align_      uint8
	FieldAlign_ uint8
	Kind_       Kind

	Equal func(unsafe.Pointer, unsafe.Pointer) bool

	GCData    *byte
	Str       NameOff
	PtrToThis TypeOff
}

type MapType struct {
	Type
	Key   *Type
	Elem  *Type
	Group *Type

	Hasher    func(unsafe.Pointer, uintptr) uintptr
	GroupSize uintptr
	SlotSize  uintptr
	ElemOff   uintptr
	Flags     uint32
}

var maptype *MapType

func (m *Map) Used() uint64 {
	return m.used
}

func (m *Map) getWithoutKey(typ *MapType, key unsafe.Pointer) (unsafe.Pointer, bool) {
	if m.Used() == 0 {
		return nil, false
	}

	if m.writing != 0 {
		log.Fatalf("concurrent map read and map write")
	}

	hash := typ.Hasher(key, m.seed)
	fmt.Println("HASH:", hash)

	if m.dirLen == 0 {
		fmt.Println("--- LITE SCENARIO ---")
		_, elem, ok := m.getWithKeySmall(typ, hash, key)
		return elem, ok
	}

	fmt.Println("--- FULL SCENARIO ---")
	idx := m.directoryIndex(hash)
	fmt.Println("Directory index:", idx)
	return m.directoryAt(idx).getWithoutKey(typ, hash, key)
}

func (m *Map) directoryIndex(hash uintptr) uintptr {
	if m.dirLen == 1 {
		return 0
	}
	return hash >> (m.globalShift & 63)
}

const PtrSize = 4 << (^uintptr(0) >> 63)

func (m *Map) directoryAt(i uintptr) *table {
	return *(**table)(unsafe.Pointer(uintptr(m.dirPtr) + PtrSize*i))
}

func (m *Map) getWithKeySmall(typ *MapType, hash uintptr, key unsafe.Pointer) (unsafe.Pointer, unsafe.Pointer, bool) {
	g := groupReference{
		data: m.dirPtr,
	}

	match := g.ctrls().matchH2(h2(hash))

	for match != 0 {
		i := match.first()

		slotKey := g.key(typ, i)
		if typ.IndirectKey() {
			slotKey = *((*unsafe.Pointer)(slotKey))
		}

		if typ.Key.Equal(key, slotKey) {
			slotElem := g.elem(typ, i)
			if typ.IndirectElem() {
				slotElem = *((*unsafe.Pointer)(slotElem))
			}
			return slotKey, slotElem, true
		}

		match = match.removeFirst()
	}

	return nil, nil, false
}

func h1(h uintptr) uintptr {
	return h >> 7
}

func h2(h uintptr) uintptr {
	return h & 0x7f
}
