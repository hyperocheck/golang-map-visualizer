package main

import (
	"visualizer/src/console"
	"visualizer/src/engine"
	"visualizer/src/preview"

	"fmt"
	"log"
	"math/rand"
	"syscall"
	"time"
)

type MyCustomData struct {
	Label bool
	Map   map[int][]string
}

func GenerateMyCustomData() MyCustomData {
	rand.Seed(time.Now().UnixNano())

	// --- Label: atomic.Bool ---
	var label bool
	if rand.Intn(2) == 1 {
		label = true
	}

	// --- Map: map[int][]string ---
	mapLen := rand.Intn(5) + 1
	m := make(map[int][]string, mapLen)

	for i := 0; i < mapLen; i++ {
		key := rand.Intn(100)
		valLen := rand.Intn(4) + 1

		values := make([]string, valLen)
		for j := 0; j < valLen; j++ {
			values[j] = randomString(5)
		}

		m[key] = values
	}

	return MyCustomData{
		Label: label,
		Map:   m,
	}
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	preview.Preview()

	usermapo := engine.Start(func(iters int, maxChain bool) map[int]int {

		m := make(map[int]int)

		for i := range 0 {
			m[i] = i
		}

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

type __noinline struct {
}
