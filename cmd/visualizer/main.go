package main

import (
	"visualizer/src/preview"
	"visualizer/src/engine"
	"visualizer/src/console"

	"log"
	"syscall"
	"fmt"
)
 
func main() {
	preview.Preview()
	
	type MyCustomData struct {
		ID    int
		Label string
		Valid bool
	}
	usermapo := engine.Start(func(iters int, maxChain bool) map[int]string {

		m := make(map[int]string)
		
		for i := range 100 {m[i] = fmt.Sprintf("%d", i) + " string"}
		
		return m
	})

	go func() {
		usermapo.PrintHmap()
		console.StartConsole(usermapo)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()
	
	if err := startServer(usermapo, ":8080"); err != nil {
		log.Printf("graceful shutdown error: %v", err)
	}
	fmt.Println("\nGoodbye!😺")
}
