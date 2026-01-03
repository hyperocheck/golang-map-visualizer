package console

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"visualizer/src/hmap"

	/* ------------------ */
	// для примера
	"visualizer/src/usermap"
	/* ------------------ */

	"github.com/fatih/color"
)

var (
	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	green  = color.New(color.FgGreen)
)

/* --------------------------------------------- */
// пример реализации сложной структуры, которая передается
// в качестве значения мапы
// испортируется из usermap

/*
type ComplexStructExmaple struct {
	i []int 
	ui uint32 
	a atomic.Bool
}

на вход подается как:
1,43,132,4,1,23,321;899;false
/* --------------------------------------------- */

func FromString(x *usermap.ComplexStructExample, input string) error {
	parts := strings.Split(input, ";")
	if len(parts) != 3 {
		return fmt.Errorf("expected 3 parts: i,ui,a")
	}

	intParts := strings.Split(parts[0], ",")
	x.I = make([]int, len(intParts))
	for idx, p := range intParts {
		v, err := strconv.Atoi(p)
		if err != nil {
			return fmt.Errorf("invalid int in slice: %v", err)
		}
		x.I[idx] = v
	}

	ui64, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		return fmt.Errorf("invalid uint32: %v", err)
	}
	x.UI = uint32(ui64)

	b, err := strconv.ParseBool(parts[2])
	if err != nil {
		return fmt.Errorf("invalid bool: %v", err)
	}
	x.A.Store(b)

	return nil
}

/* --------------------------------------------- */

func StartConsole[K comparable, V any](m map[K]V) {
	time.Sleep(150 * time.Millisecond)
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
			hmap.PrintHmap(hmap.GetHmap(m))
		case "show":
			l := len(m)
			if l > 100 {
				yellow.Printf("Больше 100 элементов. Уверен?(y/n)")
				if !scanner.Scan() {break}
				ans := scanner.Text()
				switch ans {
				case "y", "yes", "Y", "н", "Н":
					for k, v := range m {
						fmt.Printf("%v : %v\n", k, v)
					}
					continue
				default:
					continue
				}
			}
			for k, v := range m {
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
			if _, ok := m[key]; ok {
				red.Println("Key already exists")
				continue
			}
			value, err := parseStringToType[V](strings.Join(args[2:], " "))
			if err != nil {
				red.Println("Invalid value:", err)
				continue
			}
			m[key] = value
			green.Println("Inserted element successfully")
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
			if _, ok := m[key]; !ok {
				red.Println("Key does not exist")
				continue
			}
			value, err := parseStringToType[V](strings.Join(args[2:], " "))
			if err != nil {
				red.Println("Invalid value:", err)
				continue
			}
			m[key] = value
			green.Println("Updated element successfully")
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
			if _, ok := m[key]; !ok {
				red.Println("Key does not exist")
				continue
			}
			delete(m, key)
			green.Println("Deleted element successfully")
		default:
			red.Println("Unknown command:", cmd)
		}
	}
}

func parseStringToType[T any](s string) (T, error) {
	var zero T

	switch any(zero).(type) {
	case usermap.ComplexStructExample:
		u := usermap.ComplexStructExample{}
		err := FromString(&u, s)
		return any(u).(T), err	
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

