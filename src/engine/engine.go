package engine

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"reflect"
	"unsafe"

	"github.com/fatih/color"
)

type Type[K comparable, V any] struct {
	Data map[K]V
}

func Start[K comparable, V any](factory func(iters int, maxChain bool) map[K]V) *Type[K, V] {
	iters := flag.Int("range", 0, "range")
	maxChain := flag.Bool("max-chain", false, "hz cho eto budet potom pridumayou")

	flag.Parse()

	userMap := factory(*iters, *maxChain)

	return &Type[K, V]{
		Data: userMap,
	}
}

func (t *Type[K, V]) GetBucketsJSON(bucketsType string) []byte {
	h := (**Hmap)(unsafe.Pointer(&t.Data))

	if bucketsType == "oldbuckets" && (*h).oldbuckets == nil {
		return []byte("[]")
	}

	if (*h).buckets == nil {
		return []byte("[]")
	}

	bucketSize := inspectMap(t.Data)

	var bucketNum uintptr
	var b unsafe.Pointer

	if bucketsType == "oldbuckets" {
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

	allBuckets := []bucketJSON{}
	maxOverflowChainLen := 0
	bucketIDMaxOverflowChainLen := 0
	chainsNum := 0
	emptyBucketsNum := 0
	id := 0
	mainID := 0

	for i := uintptr(0); i < bucketNum; i++ {

		var new_main_bucket bucketJSON

		bucket := (*_bucket_[K, V])(unsafe.Pointer(uintptr(b) + i*bucketSize))

		new_main_bucket.Tophash = bucket.tophash
		new_main_bucket.ID = id
		new_main_bucket.Type = "main"

		if fillBucket(&new_main_bucket, bucket) {
			emptyBucketsNum++
		}
		if bucket.overflow != nil {
			new_main_bucket.Overflow = fmt.Sprintf("0x%x", bucket.overflow)
		} else {
			new_main_bucket.Overflow = "nil"
		}

		allBuckets = append(allBuckets, new_main_bucket)
		id++

		currOverflowChainLen := 0
		curr_overflow_addr := bucket.overflow
		if curr_overflow_addr != nil {
			chainsNum++
		}
		for curr_overflow_addr != nil {
			var new_overflow_bucket bucketJSON

			obucket := (*_bucket_[K, V])(unsafe.Pointer(curr_overflow_addr))

			new_overflow_bucket.Tophash = obucket.tophash
			new_overflow_bucket.ID = id
			new_overflow_bucket.Type = "overflow"

			if fillBucket(&new_overflow_bucket, obucket) {
				emptyBucketsNum++
			}
			if obucket.overflow != nil {
				new_overflow_bucket.Overflow = fmt.Sprintf("0x%x", obucket.overflow)
			} else {
				new_overflow_bucket.Overflow = "nil"
			}

			curr_overflow_addr = obucket.overflow

			allBuckets = append(allBuckets, new_overflow_bucket)
			id++
			currOverflowChainLen++
		}

		if maxOverflowChainLen < currOverflowChainLen {
			maxOverflowChainLen = currOverflowChainLen
			bucketIDMaxOverflowChainLen = mainID
		}
		mainID++
	}
	
	mapo_types := GetKVType(t)
	resp := VizualResponse{
		Buckets: allBuckets,
		Stats: BucketStats{
			LoadFactor:       float64(len(t.Data)) / float64(int(1)<<(*h).B),
			MaxChainLen:      maxOverflowChainLen,
			MaxChainBucketID: bucketIDMaxOverflowChainLen,
			NumChains:        chainsNum,
			NumEmptyBuckets:  emptyBucketsNum,
			KeyType: mapo_types[0],
			ValueType: mapo_types[1],
		},
	}

	res, err := json.Marshal(resp)
	if err != nil || len(res) == 0 {
		return []byte("[]")
	}
	return res
}

func fillBucket[K comparable, V any](b *bucketJSON, rb *_bucket_[K, V]) bool {
	emptyKeyNum := 0
	for j := 0; j < 8; j++ {
		// Если слот пустой (tophash < 5)
		if rb.tophash[j] < 5 {
			emptyKeyNum++
			b.Keys[j] = nil
			b.Values[j] = nil
		} else {
			// Маршалим ключ напрямую из структуры бакета
			if kBytes, err := json.Marshal(rb.keys[j]); err == nil {
				b.Keys[j] = json.RawMessage(kBytes)
			} else {
				// Если ошибка (маловероятно), записываем как ошибку
				b.Keys[j] = json.RawMessage(`"error marshalling"`)
			}

			// Маршалим значение
			if vBytes, err := json.Marshal(rb.values[j]); err == nil {
				b.Values[j] = json.RawMessage(vBytes)
			} else {
				b.Values[j] = json.RawMessage(`"error marshalling"`)
			}
		}
	}
	// Возвращаем true, если весь бакет пуст (все 8 слотов)
	return emptyKeyNum == 8
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
		jsonH.Extra = []string{fmt.Sprintf("%p", h.extra.overflow), fmt.Sprintf("%p", h.extra.oldoverflow), fmt.Sprintf("%p", h.extra.nextOverflow)}
	} else {
		jsonH.Extra = []string{"0x0", "0x0", "0x0"}
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

func (t *Type[K, V]) GetHmap() *Hmap {
	if t.Data == nil {
		return nil
	}
	return *(**Hmap)(unsafe.Pointer(&t.Data))
}

func (t *Type[K, V]) Generate() {

	h := (**Hmap)(unsafe.Pointer(&t.Data))

	bucketSize := inspectMap(t.Data)

	cmax := 0
	mstr := ""
	for i := uintptr(0); i < uintptr(1)<<(*h).B; i++ {
		if (*h).B == 0 {
			break
		}
		bucket := uintptr(unsafe.Pointer((*h).buckets)) + bucketSize*i
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

	log.Println("max_chain_lenght: ", cmax)
	log.Println("max_chain: ", mstr)
}

func GetKVType[K comparable, V any](t *Type[K, V]) [2]string {
	var out [2]string

	var k K
	var v V

	out[0] = reflect.TypeOf(k).String()
	out[1] = reflect.TypeOf(v).String()

	return out
}

func (t *Type[K, V]) PrintHmap() {

	h := t.GetHmap()

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
	// угарчик
	start := [3]int{180, 80, 255}
	end := [3]int{80, 200, 255}

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
		h.B, uintptr(1)<<(*h).B,
		h.noverflow,
		h.hash0,
		h.buckets,
		h.oldbuckets,
		h.nevacuate,
		h.extra,
	)
	color.Magenta("\n}")
}
