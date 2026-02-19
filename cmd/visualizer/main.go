package main

import (
	"visualizer/src/cmd"
	"visualizer/src/engine"
)

func main() {
	// Create your map here and be sure to return it.
	// You can do anything with the map inside this block.
	// And also don't forget to specify the return type.

	m := make(engine.Map[int, int], 100)

	for i := cmd.Flag.From; i < cmd.Flag.To; i++ {
		m[i] = i
	}

	work(m)
}
