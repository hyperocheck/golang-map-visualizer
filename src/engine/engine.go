package engine

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
	"visualizer/src/console"
)

type (
	Meta[K comparable, V any] struct {
		Console      *console.Console
		Map          map[K]V
		bucketSizeof uintptr
		mu           sync.Mutex
		ktype        string
		vtype        string
	}
	Map[K comparable, V any] map[K]V
)

func GetMetaByMap[K comparable, V any](t Map[K, V]) *Meta[K, V] {
	kv := GetKVType(t)

	return &Meta[K, V]{
		Map:          t,
		bucketSizeof: GetBucketSize(t),
		mu:           sync.Mutex{},
		ktype:        kv[0],
		vtype:        kv[1],
	}
}

type MapBucketType int

const (
	MapBucketNew MapBucketType = iota
	MapBucketOld
)

func (t *Meta[K, V]) GetBucketsJSON(btype MapBucketType) ([]byte, error) {
	h := GetHmap(t.Map)

	if h == nil {
		return nil, errors.New("map not initialized")
	}

	var (
		bucketCount uintptr
		buckets     unsafe.Pointer
	)

	switch btype {
	case MapBucketOld:
		if (*h).oldbuckets == nil || (*h).B == 0 {
			return []byte("[]"), nil
		}
		bucketCount = uintptr(1) << ((*h).B - 1)
		buckets = (*h).oldbuckets

	case MapBucketNew:
		if (*h).buckets == nil {
			return []byte("[]"), nil
		}
		bucketCount = uintptr(1) << (*h).B
		buckets = (*h).buckets
	}

	var (
		bucketsJSON         = make([]bucketJSON[K, V], 0, bucketCount)
		maxOverflowChainLen int
		chainsCount         int
		emptyBucketsCount   int
	)

	for i := uintptr(0); i < bucketCount; i++ {
		bucket := (*_bucket_[K, V])(unsafe.Add(buckets, i*t.bucketSizeof))

		b := bucketJSON[K, V]{
			Tophash:  bucket.tophash,
			Keys:     bucket.keys,
			Values:   bucket.values,
			Overflow: "0x" + strconv.FormatUint(uint64(uintptr(bucket.overflow)), 16),
		}

		emptySlots := 0
		for j := range 8 {
			if bucket.tophash[j] < 5 {
				emptySlots++
			}
		}
		if emptySlots == 8 {
			emptyBucketsCount++
		}

		bucketsJSON = append(bucketsJSON, b)

		currOverflowChainLen := 0
		currOverflow := bucket.overflow
		if currOverflow != nil {
			chainsCount++
		}
		for currOverflow != nil {
			bucket := (*_bucket_[K, V])(unsafe.Pointer(currOverflow))

			b := bucketJSON[K, V]{
				Tophash:  bucket.tophash,
				Keys:     bucket.keys,
				Values:   bucket.values,
				Overflow: "0x" + strconv.FormatUint(uint64(uintptr(bucket.overflow)), 16),
			}

			emptySlots := 0
			for j := range 8 {
				if bucket.tophash[j] < 5 {
					emptySlots++
				}
			}
			if emptySlots == 8 {
				emptyBucketsCount++
			}

			currOverflow = bucket.overflow
			bucketsJSON = append(bucketsJSON, b)
			currOverflowChainLen++
		}

		maxOverflowChainLen = max(maxOverflowChainLen, currOverflowChainLen)
	}

	resp := VizualResponse[K, V]{
		Buckets: bucketsJSON,
		Stats: BucketStats{
			LoadFactor:      float64(len(t.Map)) / float64(int(1)<<(*h).B),
			MaxChainLen:     maxOverflowChainLen,
			NumChains:       chainsCount,
			NumEmptyBuckets: emptyBucketsCount,
			KeyType:         t.ktype,
			ValueType:       t.vtype,
		},
	}

	res, err := json.Marshal(resp)
	if err != nil || len(res) == 0 {
		return nil, err
	}
	return res, nil
}

func GetHmapJSON(h *Hmap) ([]byte, error) {
	if h == nil {
		return []byte(`{"error":"Hmap is nil"}`), nil
	}

	NumBuckets := 0
	if h.buckets != nil {
		NumBuckets = 1 << h.B
	}
	jsonH := hmapJSON{
		Count:      h.count,
		Flags:      h.flags,
		B:          h.B,
		NumBuckets: NumBuckets,
		NOverflow:  h.noverflow,
		Hash0:      h.Hash0,
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

func GetBucketSize[K comparable, V any](m map[K]V) uintptr {
	keyType := reflect.TypeOf(*new(K))
	valType := reflect.TypeOf(*new(V))
	keySize := keyType.Size()
	valSize := valType.Size()
	ptrSize := unsafe.Sizeof(unsafe.Pointer(nil))

	bucketSize := uintptr(8) + 8*keySize + 8*valSize + ptrSize

	return bucketSize
}

func GetHmap[K comparable, V any](t Map[K, V]) *Hmap {
	if t == nil {
		return nil
	}
	return *(**Hmap)(unsafe.Pointer(&t))
}

func GetKVType[K comparable, V any](t Map[K, V]) [2]string {
	var out [2]string

	var k K
	var v V

	out[0] = reflect.TypeOf(k).String()
	out[1] = reflect.TypeOf(v).String()

	return out
}

/*
func (t Map[K, V]) PrintHmap() {
	h := GetHmap(t)

	lines := []string{
		"Hmap {",
		fmt.Sprintf("  count       %v", h.count),
		fmt.Sprintf("  flags       %v", h.flags),
		fmt.Sprintf("  B           %v", h.B),
		fmt.Sprintf("  noverflow   %v", h.noverflow),
		fmt.Sprintf("  hash0       %v", h.Hash0),
		fmt.Sprintf("  buckets     0x%x", h.buckets),
		fmt.Sprintf("  oldbuckets  0x%x", h.oldbuckets),
		fmt.Sprintf("  nevacuate   %v", h.nevacuate),
		fmt.Sprintf("  extra       %x", h.extra),
		"}",
	}

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
*/
