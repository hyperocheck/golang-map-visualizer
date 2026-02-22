package main

import (
	"visualizer/internal/flags"
	"visualizer/internal/legacy/engine"
)

func main() {
	m := make(engine.Map[int, int], 1)

	for i := flags.Flag.From; i < flags.Flag.To; i++ {
		m[i] = i
	}

	work(m)
}
