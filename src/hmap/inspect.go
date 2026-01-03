package hmap

import (
	"unsafe"
	"reflect"
	"encoding/json"
	"fmt"

	"github.com/fatih/color"
)

func GetBucketsJSON[K comparable, V any](m map[K]V, _type_ string) []byte {
	h := (**Hmap)(unsafe.Pointer(&m))

	if _type_ == "oldbuckets" && (*h).oldbuckets == nil {
		return []byte("[]")
	}

	if (*h).buckets == nil {
		return []byte("[]")
	}

	bucketSize := inspectMap(m)
	
	var bucketNum uintptr
	var b unsafe.Pointer
	
	if _type_ == "oldbuckets" {
		b = (*h).oldbuckets
		if b == nil {
			return []byte("[]")
		}
		if (*h).B == 0 {
			return []byte("[]")
		}
		bucketNum = uintptr(1) << ((*h).B - 1)
	} else {
		b = (*h).buckets
		if b == nil {
			return []byte("[]")
		}
		bucketNum = uintptr(1) << (*h).B
	}

	allBuckets := []bucketJSON[K, V]{}
	id := 0
	
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
	
    //res, err := json.MarshalIndent(allBuckets, "", "	")
	res, err := json.Marshal(allBuckets)

	if err != nil || len(res) == 0 {
		return []byte("[]")
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

func GetHmapJSON(h *Hmap) ([]byte, error) {
	if h == nil {
		return []byte(`{"error":"Hmap is nil"}`), nil
	}

	jsonH := hmapJSON{
		Count:      h.count,
		Flags:      h.flags,
		B:          h.B,
		NumBuckets: 1 << h.B,
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

func inspectMap[K comparable, V any](m map[K]V) uintptr {
	keyType := reflect.TypeOf(*new(K))
	valType := reflect.TypeOf(*new(V))
	keySize := keyType.Size()
	valSize := valType.Size()
	ptrSize := unsafe.Sizeof(unsafe.Pointer(nil)) 

	bucketSize := uintptr(8) + 8*keySize + 8*valSize + ptrSize
	
	//log.Println("map bucketSize=", bucketSize)

	return bucketSize 
}

func GetHmap[K comparable, V any](m map[K]V) *Hmap {
	if m == nil {
		return nil
	}
	return *(**Hmap)(unsafe.Pointer(&m))
}

/*
func generate[K comparable, V any](m map[K]V) {

	h := (**Hmap)(unsafe.Pointer(&m))
		
	bucketSize := inspectMap(m)

	cmax := 0
	mstr := ""
	for i := uintptr(0); i < uintptr(1) << (*h).B; i++ {
		if (*h).B == 0 {
			break
		}
		bucket := uintptr(unsafe.Pointer((*h).buckets)) + bucketSize * i
		rb := (*_bucket_[K, V])(unsafe.Pointer(bucket))
		curr := rb.overflow
		count := 0
		maxstr := ""
		for curr != nil {
			count++
			maxstr += fmt.Sprintf("%v -> ", rb.overflow)
			maxstr += fmt.Sprintf("%v - %v", (*_bucket_[K, V])(curr).keys, (*_bucket_[K, V])(curr).values)
			curr = (*_bucket_[K, V])(curr).overflow
		}
		if count > cmax {
			cmax = count 
			mstr = maxstr
		}
	}
	fmt.Println("COUNT: ", cmax)
	fmt.Println(mstr)
}
*/

func PrintHmap(h *Hmap) {
	lines := []string{
		"Hmap {",
		fmt.Sprintf("  count       %v", h.count),
		fmt.Sprintf("  flags       %v", h.flags),
		fmt.Sprintf("  B           %v -> num_buckets=%v", h.B, uintptr(1)<<h.B),
		fmt.Sprintf("  noverflow   %v", h.noverflow),
		fmt.Sprintf("  hash0       %v", h.hash0),
		fmt.Sprintf("  buckets     0x%x", h.buckets),
		fmt.Sprintf("  oldbuckets  0x%x", h.oldbuckets),
		fmt.Sprintf("  nevacuate   %v", h.nevacuate),
		fmt.Sprintf("  extra       %x", h.extra),
		"}",
	}
// унарчик
	start := [3]int{180, 80, 255}  
	end   := [3]int{80, 200, 255}

	steps := len(lines) - 1
	for i, line := range lines {
		r := start[0] + (end[0]-start[0])*i/steps
		g := start[1] + (end[1]-start[1])*i/steps
		b := start[2] + (end[2]-start[2])*i/steps

		color.RGB(r, g, b).Println(line)
	}
}

func printHmap(h *Hmap) {
	magneta := color.New(color.FgMagenta)
	
	magneta.Printf("Hmap {")
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
			h.count, 
			h.flags, 
			h.B, uintptr(1) << (*h).B, 
			h.noverflow, 
			h.hash0, 
			h.buckets,
			h.oldbuckets,
			h.nevacuate, 
			h.extra,
		)
	color.Magenta("\n}")
}





