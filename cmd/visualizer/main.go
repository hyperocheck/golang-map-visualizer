package main

import (
	"visualizer/src/console"
	"visualizer/src/hmap"
	"visualizer/src/preview"
	"visualizer/src/usermap"

	"log"
	"syscall"
	"fmt"
)

var (
	m = usermap.GetUserMap()
)

func main() {
	preview.Preview()
	
	go func() {
		hmap.PrintHmap(hmap.GetHmap(m))
		console.StartConsole(m)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()
	
	if err := startServer(":8080"); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	fmt.Println("\nGoodbye!😺")
}
