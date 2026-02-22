package main

import (
	"fmt"
	"log"
	"syscall"

	"visualizer/internal/console"
	"visualizer/internal/legacy/engine"
	"visualizer/internal/preview"
)

func work[K comparable, V any](t engine.Map[K, V]) {
	preview.Preview()

	cons := console.New()
	meta := engine.GetMetaByMap(t)
	meta.Console = cons
	meta.RegisterCommands()
	meta.Console.SetPrompt("> ")

	go func() {
		cons.Run()
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	if err := startServer(meta, ":9090"); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	fmt.Println("\nGoodbye!😺")
}
