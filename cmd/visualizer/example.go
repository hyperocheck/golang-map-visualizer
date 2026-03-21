package main

import (
	"fmt"
	"strconv"
	"strings"
)

// Stats - кастомный тип значения для демонстрации использования программы с кастомными типами
type Stats struct {
	Nums   []int `json:"nums"`
	Active bool  `json:"active"`
}

/*
Необязательно реализовывать Parse.
Тогда в команде insert для мапы [string]Stats нужно будет писать команду json'ом:
insert hello '{"nums":[1,2,3,4,5,6],"active":true}'

func (s Stats) Parse(str string) (Stats, error) {
	...
}
*/

func (s Stats) FromIndex(n int64) (Stats, bool) {
	return Stats{
		Nums:   []int{int(n), int(n * 2), int(n * 3)},
		Active: n%2 == 0,
	}, true
}

func (s Stats) String() string {
	parts := make([]string, len(s.Nums))
	for i, v := range s.Nums {
		parts[i] = strconv.Itoa(v)
	}
	active := "off"
	if s.Active {
		active = "on"
	}
	return fmt.Sprintf("[%s] active=%s", strings.Join(parts, ","), active)
}
