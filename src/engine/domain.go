package engine

import (
	"encoding/json"
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

type VizualResponse struct {
	Buckets []bucketJSON `json:"buckets"`
	Stats   BucketStats  `json:"stats"`
}

type Hmap struct {
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

type bucketJSON struct {
	Tophash  [8]uint8           `json:"tophash"`
	Keys     [8]json.RawMessage `json:"keys,omitempty"`
	Values   [8]json.RawMessage `json:"values,omitempty"`
	Overflow string             `json:"overflow"`

	Type string `json:"type"` // main || overflow
	ID   int    `json:"id"`   // main bucket id (bid)
}
