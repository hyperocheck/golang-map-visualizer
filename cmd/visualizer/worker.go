package main

import (
	"fmt"
	"log"
	"syscall"
	
	"visualizer/src/console"
	"visualizer/src/engine"
	"visualizer/src/preview"
)

func work[K comparable, V any](fn func(i_from, i_to int) map[K]V) {
	usermapo := engine.Start(fn)

	preview.Preview()
	usermapo.PrintHmap()

	go func() {
		console.StartConsole(usermapo)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	if err := startServer(usermapo, ":8080"); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	fmt.Println("\nGoodbye!😺")
}
