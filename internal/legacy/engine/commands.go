package engine

import (
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	"visualizer/internal/ws"

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
	m.registerMapAccess1()
	m.registerHelp()
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
	m.Console.RegisterCommand("evil", "evil <count> [--life] [--bid <int>] — insert keys into target bucket (map[int]int only)", func(ctx *ishell.Context) {
		if err := m.CheckTypeInt(); err != nil {
			ctx.PrintlnLogError(err)
			return
		}

		args := ctx.Args
		if len(args) < 1 {
			ctx.PrintlnLogWarn("Usage: evil <count> [--life] [--bid <int>]")
			return
		}

		count, err := strconv.Atoi(args[0])
		if err != nil || count <= 0 {
			ctx.PrintlnLogError("Invalid count:", args[0])
			return
		}

		liveMode := false
		BUCKET_NUM := uint8(0)
		flagArgs := args[1:]
		for i := 0; i < len(flagArgs); i++ {
			switch flagArgs[i] {
			case "--life":
				liveMode = true
			case "--bid":
				if i+1 >= len(flagArgs) {
					ctx.PrintlnLogWarn("--bid requires an integer argument")
					return
				}
				i++
				bid, err := strconv.Atoi(flagArgs[i])
				if err != nil || bid < 0 {
					ctx.PrintlnLogError("Invalid --bid value:", flagArgs[i])
					return
				}
				BUCKET_NUM = uint8(bid)
			}
		}
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
					ctx.PrintfLogEvent("[%d/%d] Inserted key=%v, attempts=%d, time=%v", inserted, count, probe, totalAttempts, elapsed)
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
		}
		totalAttempts = 0
		ctx.PrintfLogGood("Evil mode completed! Inserted %d keys, total attempts: %d", count, totalAttempts)

		//ctx.PrintfLogEvent("Evil mode completed!")
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
	m.Console.RegisterCommand("range", "range <insert|delete> <from> <to> [--life] — bulk operation on range of keys (map[int]int only)", func(ctx *ishell.Context) {
		if err := m.CheckTypeInt(); err != nil {
			ctx.PrintlnLog(err)
			return
		}

		args := ctx.Args
		if len(args) < 3 {
			ctx.PrintlnLogWarn("Usage: range <insert|delete> <from> <to> [--life]")
			return
		}

		op := args[0]
		if op != "insert" && op != "delete" {
			ctx.PrintlnLogError("Unknown operation:", op, "(expected insert or delete)")
			return
		}

		from, err := strconv.Atoi(args[1])
		if err != nil {
			ctx.PrintlnLogError("Invalid from:", args[1])
			return
		}

		to, err := strconv.Atoi(args[2])
		if err != nil {
			ctx.PrintlnLogError("Invalid to:", args[2])
			return
		}

		if from > to {
			ctx.PrintlnLogError("Error: from must be <= to")
			return
		}

		liveMode := false
		for _, arg := range args[3:] {
			if arg == "--life" {
				liveMode = true
				break
			}
		}

		affected := 0
		skipped := 0

		if op == "insert" {
			if liveMode {
				ctx.Printf("Range insert (live): keys from %d to %d...\n", from, to)
				for i := from; i <= to; i++ {
					key := any(i).(K)
					val := any(i).(V)
					if _, exists := m.Map[key]; exists {
						skipped++
						continue
					}
					m.Map[key] = val
					affected++
					ws.NotifyUpdate()
					time.Sleep(100 * time.Millisecond)
					ctx.PrintfLogEvent("Inserted key %v : val %v", key, val)
				}
			} else {
				ctx.PrintfLog("Range insert (batch): keys from %d to %d...", from, to)
				for i := from; i <= to; i++ {
					key := any(i).(K)
					val := any(i).(V)
					if _, exists := m.Map[key]; exists {
						skipped++
					} else {
						m.Map[key] = val
						affected++
					}
				}
				ws.NotifyUpdate()
			}
			ctx.PrintfLogGood("Range insert done! Inserted: %d, skipped (already exist): %d", affected, skipped)
		} else {
			if liveMode {
				ctx.Printf("Range delete (live): keys from %d to %d...\n", from, to)
				for i := from; i <= to; i++ {
					key := any(i).(K)
					if _, exists := m.Map[key]; !exists {
						skipped++
						continue
					}
					delete(m.Map, key)
					affected++
					ws.NotifyUpdate()
					time.Sleep(100 * time.Millisecond)
					ctx.PrintfLogEvent("Deleted key %v", key)
				}
			} else {
				ctx.PrintfLog("Range delete (batch): keys from %d to %d...", from, to)
				for i := from; i <= to; i++ {
					key := any(i).(K)
					if _, exists := m.Map[key]; exists {
						delete(m.Map, key)
						affected++
					} else {
						skipped++
					}
				}
				ws.NotifyUpdate()
			}
			ctx.PrintfLogGood("Range delete done! Deleted: %d, skipped (not found): %d", affected, skipped)
		}
	})
}

func (m *Meta[K, V]) registerShow() {
	m.Console.RegisterCommand("show", "show — print all key-value pairs", func(ctx *ishell.Context) {
		mapSize := len(m.Map)

		if mapSize > 100 {
			ctx.PrintfLogWarn_("Map contains more than 100 elements. Are you sure? (y/n): ")
			answer := ctx.ReadLine()
			answer = strings.ToLower(strings.TrimSpace(answer))

			if answer != "y" && answer != "yes" && answer != "н" {
				ctx.PrintlnLog("Cancelled")
				return
			}
		}

		if mapSize == 0 {
			ctx.PrintlnLogEvent("Map is empty")
			return
		}

		ctx.Printf("Showing %d key-value pairs:\n", mapSize)
		for k, v := range m.Map {
			ctx.Printf("%v : %v\n", k, v)
		}
		ctx.PrintfLogGood("Total: %d pairs", mapSize)
	})
}

func (m *Meta[K, V]) registerHmap() {
	m.Console.RegisterCommand("hmap", "hmap — show internal hmap structure", func(ctx *ishell.Context) {
		PrintHmap2(m.Map, ctx)
	})
}

func PrintHmap2[K comparable, V any](
	t Map[K, V],
	shell *ishell.Context,
) {
	h := GetHmap(t)

	lines := []string{
		"Hmap {",
		fmt.Sprintf("  count       %v", h.count),
		fmt.Sprintf("  flags       %v", h.flags),
		fmt.Sprintf("  B           %v", h.B),
		fmt.Sprintf("  noverflow   %v", h.noverflow),
		fmt.Sprintf("  hash0       %v", h.Hash0),
		fmt.Sprintf("  buckets     0x%x", h.buckets),
		fmt.Sprintf("  oldbuckets  0x%x", h.oldbuckets),
		fmt.Sprintf("  nevacuate   %v", h.nevacuate),
		fmt.Sprintf("  extra       %x", h.extra),
		"}",
	}

	isWindows := runtime.GOOS == "windows"

	start := [3]int{180, 80, 255}
	end := [3]int{80, 200, 255}
	steps := len(lines) - 1

	for i, line := range lines {
		var colored string

		if isWindows {
			colored = fmt.Sprintf("\x1b[36m%s\x1b[0m", line)
		} else {
			r := start[0] + (end[0]-start[0])*i/steps
			g := start[1] + (end[1]-start[1])*i/steps
			b := start[2] + (end[2]-start[2])*i/steps
			colored = fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, line)
		}

		shell.Println(colored)
	}
}

func buildHelpMessage() string {
	rgb := func(r, g, b int, s string) string {
		return fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, s)
	}
	y := func(s string) string { return rgb(255, 195, 55, s) }  // amber  — commands
	d := func(s string) string { return rgb(115, 115, 135, s) } // dim    — args & notes
	w := func(s string) string { return rgb(210, 215, 225, s) } // white  — descriptions
	tl := func(s string) string { return rgb(55, 195, 210, s) } // teal   — flags
	p := func(s string) string { return rgb(160, 120, 255, s) } // purple — section headers

	row := func(cmd, args, desc string) string {
		return fmt.Sprintf("│    %s%s%s",
			y(fmt.Sprintf("%-8s", cmd)),
			d(fmt.Sprintf("%-14s", args)),
			w(desc),
		)
	}
	flag := func(name, desc string) string {
		return fmt.Sprintf("│             %s     %s", tl("╰──➤  "+name), d(desc))
	}
	sec := func(name string) string {
		return "│  " + p("▸ "+name)
	}

	raw := []string{
		"╭────────────────────────────────────────────────────────────────────",
		"│",
		sec("Map Operations"),
		"│",
		row("insert", "<key> <value>", "Insert a key-value pair  (comparable, any)"),
		row("update", "<key> <value>", "Update a key-value pair  (comparable, any)"),
		row("delete", "<key>", "Delete a key-value pair by key  (comparable)"),
		"│",
		sec("Bulk Operations"),
		"│",
		row("range", "<insert|delete> <from> <to> ", "Insert or delete a range of keys  (int, int)"),
		flag("--life", "step-by-step live mode"),
		"│",
		row("evil", "<n>", "Collision attack: n keys into target bucket"),
		flag("--life", "step-by-step live mode"),
		flag("--bid <int>", "target bucket number (default: 0)"),
		"│",
		sec("Inspection"),
		"│",
		row("show", "", "Display all key-value pairs"),
		row("hmap", "", "Print internal hmap structure"),
		row("mapaccess1", " <key>", "Step-by-step simulation of mapaccess1"),
		"│",
		"╰─────────────────────────────────────────────────────────────────────",
	}

	sR, sG, sB := 80, 220, 100
	eR, eG, eB := 50, 155, 210
	total := len(raw)

	var buf strings.Builder
	//buf.WriteString("\n")

	for i, line := range raw {
		frac := float64(i) / float64(total-1)
		r := int(float64(sR) + frac*float64(eR-sR))
		g := int(float64(sG) + frac*float64(eG-sG))
		b := int(float64(sB) + frac*float64(eB-sB))

		if strings.ContainsRune(line, '\x1b') {
			runes := []rune(line)
			buf.WriteString(rgb(r, g, b, string(runes[:1])) + string(runes[1:]) + "\n")
		} else {
			if i == len(raw)-1 {
				buf.WriteString(rgb(r, g, b, line))
			} else {
				buf.WriteString(rgb(r, g, b, line) + "\n")
			}
		}
	}

	return buf.String()
}

func (m *Meta[K, V]) registerHelp() {
	m.Console.RegisterCommand("help", "help!", func(ctx *ishell.Context) {
		ctx.Println(buildHelpMessage())
	})
}
