package engine

import "unsafe"
import "encoding/json"

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
	Extra      []string `json:"extra,omitempty"`
	IsGrowing  bool     `json:"isgrowing"`
}

// только для мап у которых и key и value типы не ссылочные
// например map[int32]bool{} map[float64]my_struct{}, где my_struct{i1 int32, i2 int32, i3 byte}
// Зачем? оптимизация для GC: в каждом бакете есть ссыока в любом случае - это overflow
// если у нас используется типы в мапе безссылочные, то нет смысла проверять их, однако overflow этому мешает - это ведь ссылка
// поэтому бакеты в таких типах помечается как те, что проверять GC не нужно.
// но появляется проблема проёба самой ссылки overflow, поэтому их просто дублируют в отдельный массив в mapextra
// это такое компромисс между скоростью gc и доп выделяемой памяти, или сканировать все бакеты, или вынести в отдельный слайс ссылки overflow
type mapextra struct {
	overflow     *[]*bmap
	oldoverflow  *[]*bmap
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
	Overflow string             `json:"overflow"` // просто для визуализации адреса типа 0x......

	Type string `json:"type"` // main || overflow
	ID   int    `json:"id"`   // просто на всякий случай, может на фронте это будет нужно
}
