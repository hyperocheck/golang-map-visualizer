package engine

import (
	"unsafe"
)

type BucketStats struct {
	LoadFactor       float64 `json:"loadFactor"`
	MaxChainLen      int     `json:"maxChainLen"`
	MaxChainBucketID int     `json:"maxChainBucketID"`
	NumChains        int     `json:"numChains"`
	NumEmptyBuckets  int     `json:"numEmptyBuckets"`
	KeyType          string  `json:"keytype"`
	ValueType        string  `json:"valuetype"`
}

type VizualResponse[K comparable, V any] struct {
	Buckets []bucketJSON[K, V] `json:"buckets"`
	Stats   BucketStats        `json:"stats"`
}

type Hmap struct {
	count      int
	flags      uint8
	B          uint8
	noverflow  uint16
	Hash0      uint32
	buckets    unsafe.Pointer
	oldbuckets unsafe.Pointer
	nevacuate  uintptr
	extra      *mapextra
}

type hmapJSON struct {
	Count      int      `json:"count"`
	Flags      uint8    `json:"flags"`
	B          uint8    `json:"B"`
	NumBuckets int      `json:"numBuckets"`
	NOverflow  uint16   `json:"noverflow"`
	Hash0      uint32   `json:"hash0"`
	Buckets    string   `json:"buckets"`
	OldBuckets string   `json:"oldbuckets"`
	NEvacuate  uintptr  `json:"nevacuate"`
	Extra      []string `json:"extra"`
	IsGrowing  bool     `json:"isgrowing"`
}

type mapextra struct {
	overflow     *[]*bmap
	oldoverflow  *[]*bmap
	nextOverflow *bmap
}

type bmap struct {
	tophash [8]uint8
}

/*
type mapextraJSON struct {
	Overflow     []uintptr `json:"overflow,omitempty"`
	OldOverflow  []uintptr `json:"oldoverflow,omitempty"`
	NextOverflow uintptr   `json:"nextOverflow,omitempty"`
}
*/

type _bucket_[K comparable, V any] struct {
	tophash  [8]uint8
	keys     [8]K
	values   [8]V
	overflow unsafe.Pointer
}

type bucketJSON[K comparable, V any] struct {
	Tophash  [8]uint8 `json:"tophash"`
	Keys     [8]K     `json:"keys"`
	Values   [8]V     `json:"values"`
	Overflow string   `json:"overflow"`
}

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

type maptype struct {
	_type
	Key       *_type
	Elem      *_type
	Group     *_type
	Hasher    func(unsafe.Pointer, uintptr) uintptr
	GroupSize uintptr
	SlotSize  uintptr
	ElemOff   uintptr
	Flags     uint32
}
