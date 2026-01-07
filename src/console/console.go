package console

import (
	"fmt"
	"strings"
	"time"

	"visualizer/src/engine"
	"visualizer/src/logger"
	"visualizer/src/ws"

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	green  = color.New(color.FgGreen)
)

func StartConsole[K comparable, V any](t *engine.Type[K, V]) {
	time.Sleep(200 * time.Millisecond)

	green.Println("Команды: show, hmap, delete <key>, update <key> <value>, insert <key> <value>, exit")

	rl := logger.Log.RL
	defer rl.Close()

	for {
		input, err := rl.Readline()
		if err != nil {
			continue
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		rl.SaveHistory(input)

		args := strings.Fields(input)
		if len(args) == 0 {
			continue
		}

		cmd := strings.ToLower(args[0])

		switch cmd {
		case "exit":
			return

		case "hmap":
			t.PrintHmap()

		case "show":
			if len(t.Data) > 100 {
				yellow.Printf("Больше 100 элементов. Уверен?(y/n) ")
				ans, err := rl.Readline()
				if err != nil {
					continue
				}
				switch strings.ToLower(ans) {
				case "y", "yes", "н":
					for k, v := range t.Data {
						fmt.Printf("%v : %v\n", k, v)
					}
				default:
					continue
				}
			} else {
				for k, v := range t.Data {
					fmt.Printf("%v : %v\n", k, v)
				}
			}
		case "insert":
			if len(args) < 3 {
				yellow.Println("Usage: insert <key> <value>")
				continue
			}
			key, err := engine.ParseValue[K](args[1])
			if err != nil {
				red.Println("Invalid key:", err)
				continue
			}
			if _, ok := t.Data[key]; ok {
				red.Println("Key already exists")
				continue
			}
			value, err := engine.ParseValue[V](strings.Join(args[2:], " "))
			if err != nil {
				red.Println("Invalid value:", err)
				continue
			}
			t.Data[key] = value
			green.Println("Inserted element successfully")
			ws.NotifyUpdate()

		case "update":
			if len(args) < 3 {
				yellow.Println("Usage: update <key> <value>")
				continue
			}
			key, err := engine.ParseValue[K](args[1])
			if err != nil {
				red.Println("Invalid key:", err)
				continue
			}
			if _, ok := t.Data[key]; !ok {
				red.Println("Key does not exist")
				continue
			}
			value, err := engine.ParseValue[V](strings.Join(args[2:], " "))
			if err != nil {
				red.Println("Invalid value:", err)
				continue
			}
			t.Data[key] = value
			green.Println("Updated element successfully")
			ws.NotifyUpdate()

		case "delete":
			if len(args) < 2 {
				yellow.Println("Usage: delete <key>")
				continue
			}
			key, err := engine.ParseValue[K](args[1])
			if err != nil {
				red.Println("Invalid key:", err)
				continue
			}
			if _, ok := t.Data[key]; !ok {
				red.Println("Key does not exist")
				continue
			}
			delete(t.Data, key)
			green.Println("Deleted element successfully")
			ws.NotifyUpdate()

		default:
			red.Println("Unknown command:", cmd)
		}
	}
}

