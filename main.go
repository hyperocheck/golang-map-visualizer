package main

import (
	"fmt"
	"unsafe"
	"net/http"
	"log"
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
	NumBuckets  int             `json:"numBuckets.omitempty"`
	NOverflow   uint16          `json:"noverflow"`
	Hash0       uint32          `json:"hash0"`
	Buckets     string          `json:"buckets"`   
	OldBuckets  string          `json:"oldbuckets,omitempty"`
	NEvacuate   uintptr         `json:"nevacuate"`
	Extra       *mapextraJSON   `json:"extra,omitempty"`
	IsGrowing   bool            `json:"isgrowing"`
}

func get_hmap_json(h *hmap) ([]byte, error) {
	if h == nil {
		return []byte(`{"error":"hmap is nil"}`), nil
	}

	jsonH := hmapJSON{
		Count:      h.count,
		Flags:      h.flags,
		B:          h.B,
		//NumBuckets: 1 << h.B,
		NOverflow:  h.noverflow,
		Hash0:      h.hash0,
		Buckets:    fmt.Sprintf("%p", h.buckets),
		OldBuckets: fmt.Sprintf("%p", h.oldbuckets),
		NEvacuate:  h.nevacuate,
		IsGrowing:  h.oldbuckets != nil,
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

	// return json.MarshalIndent(jsonH, "", "  ")
	return json.Marshal(jsonH)
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


func inspectMap[K comparable, V any](m map[K]V) uintptr {
			
	keyType := reflect.TypeOf(*new(K))
	valType := reflect.TypeOf(*new(V))
	keySize := keyType.Size()
	valSize := valType.Size()
	ptrSize := unsafe.Sizeof(unsafe.Pointer(nil)) 

	bucketSize := uintptr(8) + 8*keySize + 8*valSize + ptrSize
	
	log.Println("map bucketSize=", bucketSize)

	return uintptr(8) + 8*keySize + 8*valSize + ptrSize 
}

func getHmap[K comparable, V any](m map[K]V) *hmap {
	if m == nil {
		return nil
	}
	return *(**hmap)(unsafe.Pointer(&m))
}

var m = make(map[int]string)
func vizual(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/vizual")
	w.Header().Set("Content-Type", "application/json")

	jsonBytes := getJSON(m)

	w.Write(jsonBytes)
}

func vizual_hmap(w http.ResponseWriter, r *http.Request) {
	fmt.Println("/hmap")
	w.Header().Set("Content-Type", "application/json")

	res, _ := get_hmap_json(getHmap(m))
	w.Write(res)
}

func main() {
	for i := 0; i < 1500; i++ {
		m[i] = "loooooSTRING 💀💀💀🦃" + fmt.Sprintf("%d", i)
	}

	//fmt.Println(string(getJSON(m)))
	generate(m)
	h, _ := get_hmap_json(getHmap(m))
	fmt.Println(string(h))
	

	mux := http.NewServeMux()
	mux.HandleFunc("/vizual", vizual)
	mux.HandleFunc("/hmap", vizual_hmap)
	mux.Handle("/", http.FileServer(http.Dir("frontend/dist")))

	fmt.Println("Сервер запущен: http://localhost:8080/vizual")
	log.Fatal(http.ListenAndServe(":8080", mux))
}


	//generate(m)
	//fmt.Println(string(getJSON(m)))
	// res, _ := get_hmap_json(getHmap(m))
	// fmt.Println(string(res))


type bucketJSON[K comparable, V any] struct {
	Tophash  [8]uint8 `json:"tophash"` 
	Keys     [8]*K     `json:"keys,omitempty`
	Values   [8]*V 	  `json:"values,omitempty`
	Overflow string   `json:"overflow"`  // просто для визуализации адреса типа 0x......

	Type     string   `json:"type"`// main || overflow
	ID       int      `json:"id"`  // просто на всякий случай, может на фронте это будет нужно
}

func getJSON[K comparable, V any](m map[K]V) []byte {
	
	h := (**hmap)(unsafe.Pointer(&m))

	bucketSize := inspectMap(m)
	bucketNum := uintptr(1) << (*h).B

	allBuckets := []bucketJSON[K, V]{}
	id := 0

	b := (*h).buckets 
	for i := uintptr(0); i < bucketNum; i++ {
		
		var new_main_bucket bucketJSON[K, V]

		bucket := (*_bucket_[K, V])(unsafe.Pointer(uintptr(b) + i * bucketSize)) 
		
		new_main_bucket.Tophash = bucket.tophash 
		new_main_bucket.ID = id
		new_main_bucket.Type = "main"
		
		fillBucket(&new_main_bucket, bucket)
		if bucket.overflow != nil {
			new_main_bucket.Overflow = fmt.Sprintf("0x%x", bucket.overflow)
		} else {
			new_main_bucket.Overflow = "nil"	
		}

		allBuckets = append(allBuckets, new_main_bucket)
		id++

		curr_overflow_addr := bucket.overflow 
		for ;curr_overflow_addr != nil; {
			var new_overflow_bucket bucketJSON[K, V]
			
			obucket := (*_bucket_[K, V])(unsafe.Pointer(curr_overflow_addr))
			
			new_overflow_bucket.Tophash = obucket.tophash 
			new_overflow_bucket.ID = id
			new_overflow_bucket.Type = "overflow"
			
			fillBucket(&new_overflow_bucket, obucket)
			if obucket.overflow != nil {
				new_overflow_bucket.Overflow = fmt.Sprintf("0x%x", obucket.overflow)
			} else {
				new_overflow_bucket.Overflow = "nil"	
			}

			curr_overflow_addr = obucket.overflow

			allBuckets = append(allBuckets, new_overflow_bucket)
			id++
		}
	}
	
//	res, err := json.MarshalIndent(allBuckets, "", "	")
	res, err := json.Marshal(allBuckets)
	//res, err := json.Marshal(map[string]any{
	//	"buckets": allBuckets,
	//})
	if err != nil {
		return []byte{}
	}

	return res
}

func fillBucket[K comparable, V any](b *bucketJSON[K, V], rb *_bucket_[K, V]) {
	for j := 0; j < 8; j++ {
		if rb.tophash[j] < 5 {
			b.Keys[j] = nil
			b.Values[j] = nil
		} else {
			kCopy := rb.keys[j]
			vCopy := rb.values[j]
			b.Keys[j] = &kCopy
			b.Values[j] = &vCopy
		}
	}
}

func generate[K comparable, V any](m map[K]V) {

	h := (**hmap)(unsafe.Pointer(&m))
		
	bucketSize := inspectMap(m)


	//numBuckets := int(1 << h.B)

	
	


	cmax := 0
	mstr := ""
	for i := uintptr(0); i < uintptr(1) << (*h).B; i++ {
		bucket := uintptr(unsafe.Pointer((*h).buckets)) + bucketSize * i
		rb := (*_bucket_[K, V])(unsafe.Pointer(bucket))
		curr := rb.overflow
		count := 0
		maxstr := ""
		for curr != nil {
			count++
			maxstr += fmt.Sprintf("%v -> ", rb.overflow)
			maxstr += fmt.Sprintf("%v - %v", (*_bucket_[K, V])(curr).keys, (*_bucket_[K, V])(curr).values)
			//fmt.Printf(maxstr)
			curr = (*_bucket_[K, V])(curr).overflow
		}
		if count > cmax {
			cmax = count 
			mstr = maxstr
		}
		//fmt.Printf("nil\n")
	}
	fmt.Println("COUNT: ", cmax)
	fmt.Println(mstr)

	//fmt.Println((*mapextra)((*h).extra).overflow)
	
	//o, _ := get_hmap_json(*h)
	//fmt.Println(string(o))
	//print_hmap(h)
	
	//buckets_P := (*[2]realBucket)(unsafe.Pointer((*h).buckets))

	//fmt.Printf("%v\n", buckets_P)
}




