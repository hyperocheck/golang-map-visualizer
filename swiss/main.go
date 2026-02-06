package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"unsafe"
)

type MapDTO struct {
	Used              uint64     `json:"used"`
	Seed              uintptr    `json:"seed"`
	DirPtr            uintptr    `json:"dirPtr"`
	DirLen            int        `json:"dirLen"`
	GlobalDepth       uint8      `json:"globalDepth"`
	GlobalShift       uint8      `json:"globalShift"`
	Writing           uint8      `json:"writing"`
	TombstonePossible bool       `json:"tombstonePossible"`
	ClearSeq          uint64     `json:"clearSeq"`
	Tables            []TableDTO `json:"tables,omitempty"`
	SmallGroup        *GroupDTO  `json:"smallGroup,omitempty"`
}

type TableDTO struct {
	Addr       uintptr    `json:"addr"`
	Index      int        `json:"index"`
	Used       uint16     `json:"used"`
	Capacity   uint16     `json:"capacity"`
	GrowthLeft uint16     `json:"growthLeft"`
	LocalDepth uint8      `json:"localDepth"`
	LengthMask uint64     `json:"lengthMask"`
	Groups     []GroupDTO `json:"groups"`
}

type GroupDTO struct {
	Ctrls []uint8   `json:"ctrls"`
	Slots []SlotDTO `json:"slots"`
}

type SlotDTO struct {
	Key   any `json:"k"`
	Value any `json:"v"`
}

func main() {
	m := map[int]string{}
	hmap := *(**Map)(unsafe.Pointer(&m))

	if maptype == nil {
		addr := GetMapType(m)
		maptype = (*MapType)(unsafe.Pointer(addr))
		fmt.Println(*maptype)
	}

	for i := range 1800 {
		m[i] = fmt.Sprintf("string %d", i)
	}

	out := MapDTO{}

	out.Used = hmap.used
	out.Seed = hmap.seed
	out.DirPtr = uintptr(hmap.dirPtr)
	out.DirLen = hmap.dirLen
	out.GlobalDepth = hmap.globalDepth
	out.GlobalShift = hmap.globalShift
	out.Writing = hmap.writing
	out.TombstonePossible = hmap.tombstonePossible
	out.ClearSeq = hmap.clearSeq

	uniq_tables := map[uintptr]struct{}{}

	if hmap.dirLen == 0 {
		g := (*group[int, string])(hmap.dirPtr)

		gOut := GroupDTO{
			Ctrls: make([]uint8, 8),
			Slots: make([]SlotDTO, 8),
		}

		ctrls := uint64(g.ctrls)
		for i := 0; i < 8; i++ {
			gOut.Ctrls[i] = uint8(ctrls >> (i * 8))
			gOut.Slots[i] = SlotDTO{
				Key:   g.slots[i].key,
				Value: g.slots[i].elem,
			}
		}

		out.SmallGroup = &gOut
	} else {
		tables := unsafe.Slice((**table)(hmap.dirPtr), hmap.dirLen)

		fmt.Println(tables)

		out.Tables = make([]TableDTO, 0, len(tables))

		for _, t := range tables {
			t_pointer := uintptr(unsafe.Pointer(t))

			tOut := TableDTO{
				Addr:       t_pointer,
				Index:      t.index,
				Used:       t.used,
				Capacity:   t.capacity,
				GrowthLeft: t.growthLeft,
				LocalDepth: t.localDepth,
				LengthMask: t.groups.lengthMask,
			}
			if _, ok := uniq_tables[t_pointer]; ok {
				out.Tables = append(out.Tables, tOut)
				continue
			}
			uniq_tables[t_pointer] = struct{}{}

			groupCount := int(t.groups.lengthMask + 1)

			groups := unsafe.Slice(
				(*group[int, string])(t.groups.data),
				groupCount,
			)

			tOut.Groups = make([]GroupDTO, 0, groupCount)

			for _, g := range groups {
				gOut := GroupDTO{
					Ctrls: make([]uint8, 8),
					Slots: make([]SlotDTO, 8),
				}

				ctrls := uint64(g.ctrls)
				for i := 0; i < 8; i++ {
					gOut.Ctrls[i] = uint8(ctrls >> (i * 8))
					gOut.Slots[i] = SlotDTO{
						Key:   g.slots[i].key,
						Value: g.slots[i].elem,
					}
				}

				tOut.Groups = append(tOut.Groups, gOut)
			}

			out.Tables = append(out.Tables, tOut)
		}
	}

	// data, err := json.MarshalIndent(out, "", " ")
	data, err := json.Marshal(out)
	if err != nil {
		log.Println(err)
		return
	}

	// fmt.Println(string(data))

	router := http.NewServeMux()

	// API endpoint
	router.HandleFunc("/data", func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("Content-Type", "application/json")
		rw.Write(data)
	})

	// Статические файлы из frontend/dist
	fs := http.FileServer(http.Dir("./frontend/dist"))
	router.Handle("/", fs)

	log.Println("Server start on http://localhost:8080")
	log.Println(http.ListenAndServe(":8080", router))
}
