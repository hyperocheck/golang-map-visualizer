package main

import "unsafe"
import "fmt"

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

func main() {

	m := map[int]int{}
	m[1] = 4444444444
	m[2] = 4444444444
	m[3] = 4444444444
	m[4] = 4444444444

	m[5] = 4444444444
	m[6] = 4444444444
	m[7] = 4444444444
	m[8] = 4444444444

	// m[9] = 4444444444 <- grow

	hmap := *(**Map)(unsafe.Pointer(&m))

	if hmap.dirLen == 0 { // smallmap opt - dirPtr is pointer to a single group
		g := (*group[int, int])(hmap.dirPtr)
		fmt.Println(g)
		
	}

	fmt.Println(hmap)
}
