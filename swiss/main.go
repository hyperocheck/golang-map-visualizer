package main

import "unsafe"
import "fmt"
import "log"



type Map struct {
	used uint64 // то же самое что count
	seed uintptr // что же это такое
	dirPtr unsafe.Pointer 
	/* directory: []uintptr [ p ]  [ p ]  [ p ]
		          			  |      |      |
							table   table  table  -> each table is a complete swiss table 
	*/
	dirLen int // 2^globalDepth
	globalDepth uint8 // То же самое что и B для количества бакетов в hmap, только кол-во указателей на таблицы (то есть кол-во таблиц)
	globalShift uint8 
	writing uint8 // типо flags
	tombstonePossible bool // надгробия 💀
	clearSeq uint64 // 
}

type ctrlGroup uint64
// A group holds abi.SwissMapGroupSlots slots (key/elem pairs) plus their
// control word.
type groupReference struct {
	// data points to the group, which is described by typ.Group and has
	// layout:
	//
	// type group struct {
	// 	ctrls ctrlGroup
	// 	slots [abi.SwissMapGroupSlots]slot
	// }
	//
	// type slot struct {
	// 	key  typ.Key
	// 	elem typ.Elem
	// }
	data unsafe.Pointer // data *typ.Group
}

type slot[K comparable, V any] struct {
	key K 
	elem V
}

type group[K comparable, V any] struct {
	ctrls ctrlGroup
	slots [8]slot[K, V]
}

// groupsReference is a wrapper type describing an array of groups stored at
// data.
type groupsReference struct {
	// data points to an array of groups. See groupReference above for the
	// definition of group.
	data unsafe.Pointer // data *[length]typ.Group

	// lengthMask is the number of groups in data minus one (note that
	// length must be a power of two). This allows computing i%length
	// quickly using bitwise AND.
	lengthMask uint64
}

type table struct {
	used uint16
	capacity uint16
	growthLeft uint16
	localDepth uint8
	index int
	groups groupsReference
}


// probing 
type probeSeq struct {
	mask   uint64 // count groups 
	offset uint64 // hash & mask 
	index  uint64 // start: 0
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
	// If popcount is fast, replace code below with return popcount(^x & (x - 1)).
	//
	// x & -x leaves only the right-most bit set in the word. Let k be the
	// index of that bit. Since only a single bit is set, the value is two
	// to the power of k. Multiplying by a power of two is equivalent to
	// left shifting, in this case by k bits. The de Bruijn (64 bit) constant
	// is such that all six bit, consecutive substrings are distinct.
	// Therefore, if we have a left shifted version of this constant we can
	// find by how many bits it was shifted by looking at which six bit
	// substring ended up at the top of the word.
	// (Knuth, volume 4, section 7.3.1)
	return int(deBruijn64tab[(x&-x)*deBruijn64>>(64-6)])
}

// Portable implementation of first.
//
// On AMD64, this is replaced with an intrisic that simply does
// TrailingZeros64. There is no need to shift as the bitset is packed.
func bitsetFirst(b bitset) uintptr {
	return uintptr(TrailingZeros64(uint64(b))) >> 3
}

type ctrl uint8

const (
	// Maximum load factor prior to growing.
	//
	// 7/8 is the same load factor used by Abseil, but Abseil defaults to
	// 16 slots per group, so they get two empty slots vs our one empty
	// slot. We may want to reevaluate if this is best for us.
	maxAvgGroupLoad = 7

	ctrlEmpty   ctrl = 0b10000000
	ctrlDeleted ctrl = 0b11111110

	bitsetLSB   = 0x0101010101010101
	bitsetMSB   = 0x8080808080808080
	bitsetEmpty = bitsetLSB * uint64(ctrlEmpty)
)

func ctrlGroupMatchH2(g ctrlGroup, h uintptr) bitset {
	// NB: This generic matching routine produces false positive matches when
	// h is 2^N and the control bytes have a seq of 2^N followed by 2^N+1. For
	// example: if ctrls==0x0302 and h=02, we'll compute v as 0x0100. When we
	// subtract off 0x0101 the first 2 bytes we'll become 0xffff and both be
	// considered matches of h. The false positive matches are not a problem,
	// just a rare inefficiency. Note that they only occur if there is a real
	// match and never occur on ctrlEmpty, or ctrlDeleted. The subsequent key
	// comparisons ensure that there is no correctness issue.
	v := uint64(g) ^ (bitsetLSB * uint64(h))
	return bitset(((v - bitsetLSB) &^ v) & bitsetMSB)
}

// matchH2 returns the set of slots which are full and for which the 7-bit hash
// matches the given value. May return false positives.
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
	// TODO(prattmic): Do something here about truncation on cast to
	// uintptr on 32-bit systems?
	offset := uintptr(i) * typ.GroupSize

	return groupReference{
		data: unsafe.Pointer(uintptr(g.data) + offset),
	}
}

// Flag values
const (
	MapNeedKeyUpdate = 1 << iota
	MapHashMightPanic
	MapIndirectKey
	MapIndirectElem
)


func (mt *MapType) IndirectKey() bool { // store ptr to key instead of key itself
	return mt.Flags&MapIndirectKey != 0
}

func (mt *MapType) IndirectElem() bool { // store ptr to elem instead of elem itself
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
	// An empty slot is   1000 0000
	// A deleted slot is  1111 1110
	// A full slot is     0??? ????
	//
	// A slot is empty iff bit 7 is set and bit 1 is not. We could select any
	// of the other bits here (e.g. v << 1 would also work).
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
			// Finding an empty slot means we've reached the end of
			// the probe sequence.
			return nil, false
		}
	}
}
 

//go:noescape
func GetMapType(m interface{}) uintptr
type TFlag uint8
type Kind uint8
type NameOff int32
type TypeOff int32

type Type struct {
	Size_       uintptr
	PtrBytes    uintptr // number of (prefix) bytes in the type that can contain pointers
	Hash        uint32  // hash of type; avoids computation in hash tables
	TFlag       TFlag   // extra type information flags
	Align_      uint8   // alignment of variable with this type
	FieldAlign_ uint8   // alignment of struct field with this type
	Kind_       Kind    // what kind of type this is (string, int, ...)
	// function for comparing objects of this type
	// (ptr to object A, ptr to object B) -> ==?
	Equal func(unsafe.Pointer, unsafe.Pointer) bool
	// GCData stores the GC type data for the garbage collector.
	// Normally, GCData points to a bitmask that describes the
	// ptr/nonptr fields of the type. The bitmask will have at
	// least PtrBytes/ptrSize bits.
	// If the TFlagGCMaskOnDemand bit is set, GCData is instead a
	// **byte and the pointer to the bitmask is one dereference away.
	// The runtime will build the bitmask if needed.
	// (See runtime/type.go:getGCMask.)
	// Note: multiple types may have the same value of GCData,
	// including when TFlagGCMaskOnDemand is set. The types will, of course,
	// have the same pointer layout (but not necessarily the same size).
	GCData    *byte
	Str       NameOff // string form
	PtrToThis TypeOff // type for pointer to this type, may be zero
}

type MapType struct {
	Type
	Key   *Type
	Elem  *Type
	Group *Type // internal type representing a slot group
	// function for hashing keys (ptr to key, seed) -> hash
	Hasher    func(unsafe.Pointer, uintptr) uintptr
	GroupSize uintptr // == Group.Size_
	SlotSize  uintptr // size of key/elem slot
	ElemOff   uintptr // offset of elem in key/elem slot
	Flags     uint32
}

var (
	maptype *MapType	
)

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

	// No match here means key is not in the map.
	// (A single group means no need to probe or check for empty).
	return nil, nil, false
}


func main() {

	m := map[int]int{}
	hmap := *(**Map)(unsafe.Pointer(&m))

	if maptype == nil {
		addr := GetMapType(m)
		maptype = (*MapType)(unsafe.Pointer(addr))
		fmt.Println(maptype)
	}
	
	for i := range 850 {m[i] = i} 
	
	key := 578
	p, ok := hmap.getWithoutKey(maptype, unsafe.Pointer(&key))
	fmt.Println(p, ok)
	
	/*
	if hmap.dirLen == 0 {
		fmt.Println("--- Small Map Optimization ---")
		g := (*group[int, int])(hmap.dirPtr)
		printGroup(g)
	} else {
		fmt.Printf("--- Directory Mode (len: %d) ---\n", hmap.dirLen)
		
		tables := unsafe.Slice((**table)(hmap.dirPtr), hmap.dirLen)
		seenTables := make(map[*table]bool)

		for i, tbl := range tables {
			if tbl == nil || seenTables[tbl] {
				continue
			}
			seenTables[tbl] = true
			fmt.Printf("table #%d (addr: %p, used: %d):\n", i, tbl, tbl.used)


			numGroups := int(tbl.groups.lengthMask + 1)
			
			groups := unsafe.Slice((*group[int, int])(tbl.groups.data), numGroups)

			for gIdx, g := range groups {
				fmt.Printf("  -- Group %d --\n", gIdx)
    			printGroup(&g) 
			}
		}
	}
	*/
}

func printGroup[K comparable, V any](g *group[K, V]) {
	ctrlBytes := (*[8]uint8)(unsafe.Pointer(&g.ctrls))

	for i := 0; i < 8; i++ {
		ctrl := ctrlBytes[i]
		
		if ctrl == 0x80 {
			fmt.Printf("  slot %d: empty\n", i)
		} else if ctrl == 0xFE {
			fmt.Printf("  slot %d: tombstone\n", i)
		} else {
			s := g.slots[i]
			fmt.Printf("  slot %d: key: %v, val: %v (ctrl: %02x)\n", i, s.key, s.elem, ctrl)
		}
	}
}

// Extracts the H1 portion of a hash: the 57 upper bits.
// TODO(prattmic): what about 32-bit systems?
func h1(h uintptr) uintptr {
	return h >> 7
}

// Extracts the H2 portion of a hash: the 7 bits not used for h1.
//
// These are used as an occupied control byte.
func h2(h uintptr) uintptr {
	return h & 0x7f
}

