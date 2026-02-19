package main

import (
	"fmt"
	"log"

	"syscall"

	"visualizer/src/console"
	"visualizer/src/engine"
	"visualizer/src/preview"
)

func work[K comparable, V any](t engine.Map[K, V]) {
	preview.Preview()

	cons := console.New()
	meta := engine.GetMetaByMap(t)
	meta.Console = cons
	meta.RegisterCommands()

	go func() {
		cons.Run()
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	if err := startServer(meta, ":8080"); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	fmt.Println("\nGoodbye!😺")
}
