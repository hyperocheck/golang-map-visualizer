package main

import (
	"visualizer/src/console"
	"visualizer/src/engine"
	"visualizer/src/preview"

	"fmt"
	"log"
	"math/rand"
	"time"
	"syscall"
	"strings"
	"strconv"
)

func main() {
	preview.Preview()

	usermapo := engine.Start(func(iters int, maxChain bool) map[int]int {

		m := make(map[int]int)

		for i := range 25 {
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

func (MyCustomData) Parse(s string) (MyCustomData, error) {
	var result MyCustomData

	s = strings.TrimSpace(s)

	// split bool ; map
	parts := strings.SplitN(s, ";", 2)
	if len(parts) != 2 {
		return result, fmt.Errorf("expected format: <bool>;<int>:{a,b,c}")
	}

	// --- parse bool ---
	label, err := strconv.ParseBool(parts[0])
	if err != nil {
		return result, fmt.Errorf("invalid bool: %w", err)
	}

	// --- parse map ---
	mapPart := strings.TrimSpace(parts[1])

	// ожидаем: <int>:{a,b,c}
	colon := strings.Index(mapPart, ":")
	if colon == -1 {
		return result, fmt.Errorf("expected map format: <int>:{a,b,c}")
	}

	// key
	keyStr := strings.TrimSpace(mapPart[:colon])
	key, err := strconv.Atoi(keyStr)
	if err != nil {
		return result, fmt.Errorf("invalid map key: %w", err)
	}

	// value part
	valPart := strings.TrimSpace(mapPart[colon+1:])
	if !strings.HasPrefix(valPart, "{") || !strings.HasSuffix(valPart, "}") {
		return result, fmt.Errorf("map values must be in { }")
	}

	valBody := valPart[1 : len(valPart)-1]
	values := []string{}

	if strings.TrimSpace(valBody) != "" {
		rawItems := strings.Split(valBody, ",")
		for _, item := range rawItems {
			values = append(values, strings.TrimSpace(item))
		}
	}

	result.Label = label
	result.Map = map[int][]string{
		key: values,
	}

	return result, nil
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
