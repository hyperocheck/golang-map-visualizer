package main

import (
	"visualizer/internal/legacy/engine"
)

func main() {
	m := make(engine.Map[string, int])

	// m - обычная мапа, с ней точно такое же взаимодействие
	// здесь можно с ней предварительно сделать что угодно

	/*
		for i := flags.Flag.From; i < flags.Flag.To; i++ {
			// m[i] = i
		}
	*/

	work(m) // не удалять :)
}
