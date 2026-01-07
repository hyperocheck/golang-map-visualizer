package console

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"visualizer/src/engine"
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
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		args := strings.Fields(line)
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
			l := len(t.Data)
			if l > 100 {
				yellow.Printf("Больше 100 элементов. Уверен?(y/n)")
				if !scanner.Scan() {
					break
				}
				ans := scanner.Text()
				switch ans {
				case "y", "yes", "Y", "н", "Н":
					for k, v := range t.Data {
						fmt.Printf("%v : %v\n", k, v)
					}
					continue
				default:
					continue
				}
			}
			for k, v := range t.Data {
				fmt.Printf("%v : %v\n", k, v)
			}

		case "insert":
			if len(args) < 3 {
				yellow.Println("Usage: insert <key> <value>")
				continue
			}
			key, err := parseStringToType[K](args[1])
			if err != nil {
				red.Println("Invalid key:", err)
				continue
			}
			if _, ok := t.Data[key]; ok {
				red.Println("Key already exists")
				continue
			}
			value, err := parseStringToType[V](strings.Join(args[2:], " "))
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
			key, err := parseStringToType[K](args[1])
			if err != nil {
				red.Println("Invalid key:", err)
				continue
			}
			if _, ok := t.Data[key]; !ok {
				red.Println("Key does not exist")
				continue
			}
			value, err := parseStringToType[V](strings.Join(args[2:], " "))
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
			key, err := parseStringToType[K](args[1])
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

func parseStringToType[T any](s string) (T, error) {
	var zero T

	switch any(zero).(type) {

	case string:
		return any(s).(T), nil
	case int:
		v, err := strconv.Atoi(s)
		return any(v).(T), err
	case int8:
		v, err := strconv.ParseInt(s, 10, 8)
		return any(int8(v)).(T), err
	case int16:
		v, err := strconv.ParseInt(s, 10, 16)
		return any(int16(v)).(T), err
	case int32:
		v, err := strconv.ParseInt(s, 10, 32)
		return any(int32(v)).(T), err
	case int64:
		v, err := strconv.ParseInt(s, 10, 64)
		return any(v).(T), err
	case uint:
		v, err := strconv.ParseUint(s, 10, 64)
		return any(uint(v)).(T), err
	case uint8:
		v, err := strconv.ParseUint(s, 10, 8)
		return any(uint8(v)).(T), err
	case uint16:
		v, err := strconv.ParseUint(s, 10, 16)
		return any(uint16(v)).(T), err
	case uint32:
		v, err := strconv.ParseUint(s, 10, 32)
		return any(uint32(v)).(T), err
	case uint64:
		v, err := strconv.ParseUint(s, 10, 64)
		return any(v).(T), err
	case float32:
		v, err := strconv.ParseFloat(s, 32)
		return any(float32(v)).(T), err
	case float64:
		v, err := strconv.ParseFloat(s, 64)
		return any(v).(T), err
	case bool:
		v, err := strconv.ParseBool(s)
		return any(v).(T), err
	default:
		return zero, fmt.Errorf("unsupported type")
	}
}


