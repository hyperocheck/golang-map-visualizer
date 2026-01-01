package main

import (
	"fmt"
	"unsafe"
	"reflect"
	"encoding/json"
	"github.com/fatih/color"
)

type hmap struct {
	count      int
	flags      uint8
	B          uint8
	noverflow  uint16
	hash0      uint32
	buckets    unsafe.Pointer
	oldbuckets unsafe.Pointer
	nevacuate  uintptr
	extra      *mapextra
}

type hmapJSON struct {
	Count       int             `json:"count"`
	Flags       uint8           `json:"flags"`
	B           uint8           `json:"B"`
	NumBuckets  int             `json:"numBuckets"`
	NOverflow   uint16          `json:"noverflow"`
	Hash0       uint32          `json:"hash0"`
	Buckets     string          `json:"buckets"`   
	OldBuckets  string          `json:"oldbuckets,omitempty"`
	NEvacuate   uintptr         `json:"nevacuate"`
	Extra       *mapextraJSON   `json:"extra,omitempty"`
	IsGrowing   bool            `json:"isGrowing,omitempty"`   // oldbuckets != nil -> true
}

func get_hmap_json(h *hmap) ([]byte, error) {
	if h == nil {
		return []byte(`{"error":"hmap is nil"}`), nil
	}

	jsonH := hmapJSON{
		Count:      h.count,
		Flags:      h.flags,
		B:          h.B,
		NumBuckets: 1 << h.B,
		NOverflow:  h.noverflow,
		Hash0:      h.hash0,
		Buckets:    fmt.Sprintf("%p", h.buckets),
		NEvacuate:  h.nevacuate,
		IsGrowing:  h.oldbuckets != nil,
	}

	if h.oldbuckets != nil {
		jsonH.OldBuckets = fmt.Sprintf("%p", h.oldbuckets)
	}

	if h.extra != nil {
		extraJSON := mapextraJSON{}

		if h.extra.nextOverflow != nil {
			extraJSON.NextOverflow = uintptr(unsafe.Pointer(h.extra.nextOverflow))
		}

		if h.extra.overflow != nil {
			slice := *h.extra.overflow
			addrs := make([]uintptr, len(slice))
			for i, b := range slice {
				if b != nil {
					addrs[i] = uintptr(unsafe.Pointer(b))
				}
			}
			extraJSON.Overflow = addrs
		}

		if h.extra.oldoverflow != nil {
			slice := *h.extra.oldoverflow
			addrs := make([]uintptr, len(slice))
			for i, b := range slice {
				if b != nil {
					addrs[i] = uintptr(unsafe.Pointer(b))
				}
			}
			extraJSON.OldOverflow = addrs
		}

		jsonH.Extra = &extraJSON
	}

	// Красивый JSON с отступами (можно убрать Indent в проде для экономии трафика)
	return json.MarshalIndent(jsonH, "", "  ")
}

// только для мап у которых и key и value типы не ссылочные
// например map[int32]bool{} map[float64]my_struct{}, где my_struct{i1 int32, i2 int32, i3 byte}
// Зачем? оптимизация для GC: в каждом бакете есть ссыока в любом случае - это overflow 
// если у нас используется типы в мапе безссылочные, то нет смысла проверять их, однако overflow этому мешает - это ведь ссылка
// поэтому бакеты в таких типах помечается как те, что проверять GC не нужно.
// но появляется проблема проёба самой ссылки overflow, поэтому их просто дублируют в отдельный массив в mapextra 
// это такое компромисс между скоростью gc и доп выделяемой памяти, или сканировать все бакеты, или вынести в отдельный слайс ссылки overflow
type mapextra struct {
	overflow    *[]*bmap
	oldoverflow *[]*bmap
	nextOverflow *bmap // это пул свободных оверфлоу бакетов, чтобы можно было быстро их взять и переиспользовать
}

type mapextraJSON struct {
	Overflow     []uintptr `json:"overflow,omitempty"`    
	OldOverflow  []uintptr `json:"oldoverflow,omitempty"` 
	NextOverflow uintptr   `json:"nextOverflow,omitempty"`
}

type bmap struct {
	tophash [8]uint8
}

func print_hmap(h **hmap) {	
	magneta := color.New(color.FgMagenta)
	
	magneta.Printf("hmap {")
	fmt.Printf(`
	count       %v 
	flags       %v
	B           %v -> num_buckets=%v 
	noverflow   %v 
	hash0       %v
	buckets     0x%x
	oldbuckets  0x%x 
	nevacuate   %v
	extra       %x`, 
			(*h).count, 
			(*h).flags, 
			(*h).B, uintptr(1) << (*h).B, 
			(*h).noverflow, 
			(*h).hash0, 
			(*h).buckets,
			(*h).oldbuckets,
			(*h).nevacuate, 
			(*h).extra,
		)
	color.Magenta("\n}")
}

type _bucket_[K comparable, V any] struct {
	tophash  [8]uint8
	keys     [8]K
	values   [8]V
	overflow unsafe.Pointer
}

func inspectMap[K comparable, V any](m map[K]V) _bucket_[K, V] {
			
//	h := *(**hmap)(unsafe.Pointer(&m))

	var new_bucket _bucket_[K, V]

	keyType := reflect.TypeOf(*new(K))
	valType := reflect.TypeOf(*new(V))
	keySize := keyType.Size()
	valSize := valType.Size()
	ptrSize := unsafe.Sizeof(unsafe.Pointer(nil)) // 8 на 64-bit

	bucketSize := uintptr(8) + 8*keySize + 8*valSize + ptrSize
	
	fmt.Println(keyType, valType, keySize, valSize, ptrSize, bucketSize)

	return new_bucket
}

func main() {

	type s struct {
		vec []int 
		b bool 
		i any
	}

	m := make(map[int]s)
	for i := range 100000 {
		m[i] = s {
			vec: []int{i},
			b: true,
			i: struct{}{},
		}
	}

	generate(m)
}


func generate[K comparable, V any](m map[K]V) {
	type __noinline struct {
		i1 uint64
		i2 uint64
	}

	h := (**hmap)(unsafe.Pointer(&m))
//	num_buckets := uintptr(1) << (*h).B
		
	rbucket := inspectMap(m)

	bucketSize := unsafe.Sizeof(rbucket)
	fmt.Printf("\n\nbucketSize: %d byte\n\n", bucketSize)
	
	cmax := 0
	mstr := ""
	for i := uintptr(0); i < uintptr(1) << (*h).B; i++ {
		bucket := uintptr(unsafe.Pointer((*h).buckets)) + unsafe.Sizeof(rbucket) * i
		rb := (*_bucket_[K, V])(unsafe.Pointer(bucket))
		curr := rb.overflow
		count := 0
		maxstr := ""
		for curr != nil {
			count++
			maxstr += fmt.Sprintf("%v -> ", rb.overflow)
			maxstr += fmt.Sprintf("%v - %v", (*_bucket_[K, V])(curr).keys, (*_bucket_[K, V])(curr).values)
			// fmt.Printf("%v -> ", rb.overflow)
			// fmt.Printf("%v - %v", (*realBucket)(curr).keys, (*realBucket)(curr).values)
			fmt.Printf(maxstr)
			curr = (*_bucket_[K, V])(curr).overflow
		}
		if count > cmax {
			cmax = count 
			mstr = maxstr
		}
		fmt.Printf("nil\n")
	}
	println(cmax)
	println(mstr)

	fmt.Println((*mapextra)((*h).extra).overflow)
	
	o, _ := get_hmap_json(*h)
	fmt.Println(string(o))
	//print_hmap(h)
	
	//buckets_P := (*[2]realBucket)(unsafe.Pointer((*h).buckets))

	//fmt.Printf("%v\n", buckets_P)
}




