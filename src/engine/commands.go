package engine

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"visualizer/src/ws"

	"github.com/abiosoft/ishell/v2"
)

func (m *Meta[K, V]) RegisterCommands() {
	if m.Console == nil {
		panic("Console is not initialized in Meta")
	}

	m.registerPing()
	m.registerInsert()
	m.registerUpdate()
	m.registerDelete()
	m.registerEvil()
	m.registerRange()
	m.registerShow()
	m.registerHmap()
}

func (m *Meta[K, V]) registerPing() {
	m.Console.RegisterCommand("ping", "проверка работы", func(ctx *ishell.Context) {
		ctx.Println("pong!")
	})
}

func (m *Meta[K, V]) registerInsert() {
	m.Console.RegisterCommand("insert", "insert <key> <value>", func(ctx *ishell.Context) {
		args := ctx.Args
		if len(args) < 2 {
			ctx.PrintlnLogWarn("Usage: insert <key> <value>")
			return
		}

		key, err := ParseValue[K](args[0])
		if err != nil {
			ctx.PrintlnLogError("Invalid key:", err)
			return
		}

		if _, ok := m.Map[key]; ok {
			ctx.PrintlnLogWarn("Key already exists")
			return
		}

		value, err := ParseValue[V](strings.Join(args[1:], " "))
		if err != nil {
			ctx.PrintlnLogError("Invalid value:", err)
			return
		}

		m.Map[key] = value
		ws.NotifyUpdate()
		ctx.PrintlnLogGood("Inserted successfully")
	})
}

func (m *Meta[K, V]) registerUpdate() {
	m.Console.RegisterCommand("update", "update <key> <value>", func(ctx *ishell.Context) {
		args := ctx.Args
		if len(args) < 2 {
			ctx.PrintlnLogWarn("Usage: update <key> <value>")
			return
		}

		key, err := ParseValue[K](args[0])
		if err != nil {
			ctx.PrintlnLogError("Invalid key:", err)
			return
		}

		if _, ok := m.Map[key]; !ok {
			ctx.PrintlnLogWarn("Key does not exist")
			return
		}

		value, err := ParseValue[V](strings.Join(args[1:], " "))
		if err != nil {
			ctx.PrintlnLogError("Invalid value:", err)
			return
		}

		m.Map[key] = value
		ws.NotifyUpdate()
		ctx.PrintlnLogGood("Updated successfully")
	})
}

func (m *Meta[K, V]) registerDelete() {
	m.Console.RegisterCommand("delete", "delete <key>", func(ctx *ishell.Context) {
		args := ctx.Args
		if len(args) < 1 {
			ctx.PrintlnLogWarn("Usage: delete <key>")
			return
		}

		key, err := ParseValue[K](args[0])
		if err != nil {
			ctx.PrintlnLogError("Invalid key:", err)
			return
		}

		if _, ok := m.Map[key]; !ok {
			ctx.PrintlnLogWarn("Key does not exist")
			return
		}

		delete(m.Map, key)
		ws.NotifyUpdate()
		ctx.PrintlnLogGood("Deleted successfully")
	})
}

func (m *Meta[K, V]) registerEvil() {
	m.Console.RegisterCommand("evil", "evil <count> [--life] — insert keys into bucket 0 (map[int]int only)", func(ctx *ishell.Context) {

		if err := m.CheckTypeInt(); err != nil {
			ctx.PrintlnLog(err)
			return
		}

		args := ctx.Args
		if len(args) < 1 {
			ctx.PrintlnLogWarn("Usage: evil <count> [--life]")
			return
		}

		count, err := strconv.Atoi(args[0])
		if err != nil || count <= 0 {
			ctx.PrintlnLogError("Invalid count:", args[0])
			return
		}

		liveMode := false
		for _, arg := range args[1:] {
			if arg == "--life" {
				liveMode = true
				break
			}
		}

		const BUCKET_NUM = uint8(0)
		probe := 0
		inserted := 0
		totalAttempts := 0

		if liveMode {
			ctx.Println("Evil mode (live): inserting", count, "keys with live updates...")
			for inserted < count {
				key := any(probe).(K)
				val := any(probe).(V)
				totalAttempts++

				if BUCKET_NUM == CheckHash(m, key) {
					if _, ok := m.Map[key]; ok {
						probe++
						continue
					}
					start := time.Now()
					m.Map[key] = val
					elapsed := time.Since(start)
					inserted++
					ctx.Printf("[%d/%d] Inserted key=%v, attempts=%d, time=%v\n", inserted, count, probe, totalAttempts, elapsed)
					ws.NotifyUpdate()
					totalAttempts = 0
					time.Sleep(500 * time.Millisecond)
				}
				probe++
			}
		} else {
			ctx.Println("Evil mode (batch): inserting", count, "keys...")
			for inserted < count {
				key := any(probe).(K)
				val := any(probe).(V)
				totalAttempts++

				if BUCKET_NUM == CheckHash(m, key) {
					if _, ok := m.Map[key]; !ok {
						m.Map[key] = val
						inserted++
					}
				}
				probe++
			}
			ws.NotifyUpdate()
			ctx.Printf("Inserted %d keys, total attempts: %d\n", count, totalAttempts)
		}

		ctx.PrintlnLogGood("Evil mode completed!")
	})
}

func (m *Meta[K, V]) CheckTypeInt() error {
	if m.ktype != "int" {
		return fmt.Errorf("Range command works only with map[int]int!")
	}
	if m.vtype != "int" {
		return fmt.Errorf("Range command works only with map[int]int!")
	}

	return nil
}

func (m *Meta[K, V]) registerRange() {
	m.Console.RegisterCommand("range", "range <from> <to> [--life] — insert range of keys (map[int]int only)", func(ctx *ishell.Context) {

		if err := m.CheckTypeInt(); err != nil {
			ctx.PrintlnLog(err)
			return
		}

		args := ctx.Args
		if len(args) < 2 {
			ctx.PrintlnLogWarn("Usage: range <from> <to> [--life]")
			return
		}

		from, err := strconv.Atoi(args[0])
		if err != nil {
			ctx.PrintlnLogError("Invalid from:", args[0])
			return
		}

		to, err := strconv.Atoi(args[1])
		if err != nil {
			ctx.PrintlnLogError("Invalid to:", args[1])
			return
		}

		if from > to {
			ctx.PrintlnLogError("Error: from must be <= to")
			return
		}

		liveMode := false
		for _, arg := range args[2:] {
			if arg == "--life" {
				liveMode = true
				break
			}
		}

		inserted := 0
		updated := 0

		if liveMode {
			ctx.Printf("Range mode (live): inserting keys from %d to %d with live updates...\n", from, to)
			for i := from; i <= to; i++ {
				key := any(i).(K)
				val := any(i).(V)

				if _, exists := m.Map[key]; exists {
					updated++
				} else {
					inserted++
				}
				m.Map[key] = val

				ws.NotifyUpdate()
				time.Sleep(100 * time.Millisecond)

				if (i-from+1)%10 == 0 || i == to {
					ctx.Printf("Progress: %d/%d (inserted: %d, updated: %d)\n", i-from+1, to-from+1, inserted, updated)
				}
			}
		} else {
			ctx.PrintlnLog(fmt.Sprintf("Range mode (batch): inserting keys from %d to %d...", from, to))
			for i := from; i <= to; i++ {
				key := any(i).(K)
				val := any(i).(V)

				if _, exists := m.Map[key]; exists {
					updated++
				} else {
					inserted++
				}
				m.Map[key] = val
			}
			ws.NotifyUpdate()
		}

		ctx.PrintlnLogGood(fmt.Sprintf("Range completed! Inserted: %d new keys, Updated: %d existing keys", inserted, updated))
	})
}

func (m *Meta[K, V]) registerShow() {
	m.Console.RegisterCommand("show", "show — print all key-value pairs", func(ctx *ishell.Context) {
		mapSize := len(m.Map)

		if mapSize > 100 {
			ctx.Print("Map contains more than 100 elements. Are you sure? (y/n): ")
			answer := ctx.ReadLine()
			answer = strings.ToLower(strings.TrimSpace(answer))

			if answer != "y" && answer != "yes" && answer != "н" {
				ctx.Println("Cancelled")
				return
			}
		}

		if mapSize == 0 {
			ctx.Println("Map is empty")
			return
		}

		ctx.Printf("Showing %d key-value pairs:\n", mapSize)
		for k, v := range m.Map {
			ctx.Printf("%v : %v\n", k, v)
		}
		ctx.Printf("Total: %d pairs\n", mapSize)
	})
}

func (m *Meta[K, V]) registerHmap() {
	m.Console.RegisterCommand("hmap", "hmap — show internal hmap structure", func(ctx *ishell.Context) {
		PrintHmap2(m.Map, ctx)
	})
}
