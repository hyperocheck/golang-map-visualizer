package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func main() {

	// Create your map here and be sure to return it.
	// You can do anything with the map inside this block.
	// And also don't forget to specify the return type.
	fn := func(i_from, i_to int) map[int]*int {  // <- TYPE

		m := make(map[int]*int)

		for i := i_from; i < i_to; i++ {
			m[i] = &i 
		}

		return m // <- RETURN MAP
	}
	// ------------------------------------

	work(fn)
}

// if you want to use a complex type of key and/or value, you need to implement its Parse method. 
// YOU DON'T NEED TO IMPLEMENT ANYTHING IF WE USE ONLY BASIC TYPES
// In other words, you need to determine the input format in the cli yourself to insert/update/delete it.
// For example, I want to use the int type as the map key, and the UserCustomDataExample type as the value.
// So: map[int]UserCustomDataExample 
type UserCustomDataExample struct {
	I1 int
	I2 int
	S1 []bool
}

// You need to know how to type this into the command prompt. (The author can't guess this for you, right? :))
// For example, you decided that you want to enter this type in the cli as <int>,<int>;<bool, bool, bool, ...>
// I just sent the chatgpt an example of my structure and told him how I wanted to enter it, and he wrote me this: 
// That's all, now the visualization will be with your type
func (UserCustomDataExample) Parse(s string) (UserCustomDataExample, error) {
	var result UserCustomDataExample

	s = strings.TrimSpace(s)

	// split ints ; bools
	parts := strings.SplitN(s, ";", 2)
	if len(parts) != 2 {
		return result, fmt.Errorf("expected format: <i1>,<i2>;<bool,bool,...>")
	}

	// --- parse ints ---
	intPart := strings.Split(parts[0], ",")
	if len(intPart) != 2 {
		return result, fmt.Errorf("expected two ints: <i1>,<i2>")
	}

	i1, err := strconv.Atoi(strings.TrimSpace(intPart[0]))
	if err != nil {
		return result, fmt.Errorf("invalid i1: %w", err)
	}

	i2, err := strconv.Atoi(strings.TrimSpace(intPart[1]))
	if err != nil {
		return result, fmt.Errorf("invalid i2: %w", err)
	}

	// --- parse bool slice ---
	boolPart := strings.TrimSpace(parts[1])
	boolStrs := []string{}
	if boolPart != "" {
		boolStrs = strings.Split(boolPart, ",")
	}

	s1 := make([]bool, len(boolStrs))
	for i, v := range boolStrs {
		b, err := strconv.ParseBool(strings.TrimSpace(v))
		if err != nil {
			return result, fmt.Errorf("invalid bool at index %d: %w", i, err)
		}
		s1[i] = b
	}

	result.I1 = i1
	result.I2 = i2
	result.S1 = s1

	return result, nil
}
 
// just for fun
func GenerateUserCustomDataExample() UserCustomDataExample {
	rand.Seed(time.Now().UnixNano())

	i1 := rand.Intn(100)
	i2 := rand.Intn(100)

	l := rand.Intn(5) + 1
	s1 := make([]bool, l)
	for i := 0; i < l; i++ {
		s1[i] = rand.Intn(2) == 1
	}

	return UserCustomDataExample {
		I1: i1,
		I2: i2,
		S1: s1,
	}
}

