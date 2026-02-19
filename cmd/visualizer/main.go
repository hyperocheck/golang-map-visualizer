package main

import (
	"visualizer/src/cmd"
	"visualizer/src/engine"
)

func main() {
	m := make(engine.Map[int, int], 1)

	for i := cmd.Flag.From; i < cmd.Flag.To; i++ {
		m[i] = i
	}

	work(m)
}
