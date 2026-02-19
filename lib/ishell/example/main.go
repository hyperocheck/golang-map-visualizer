package main

import (
	"fmt"
	"runtime"
	"strings"
	"time"
	"unsafe"

	"github.com/abiosoft/ishell/v2"
	//"github.com/fatih/color"
	// "github.com/abiosoft/readline"
)

type mapextra struct {
	overflow     *[]*bmap
	oldoverflow  *[]*bmap
	nextOverflow *bmap
}

type bmap struct {
	tophash [8]uint8
}

type Hmap struct {
	count      int
	flags      uint8
	B          uint8
	noverflow  uint16
	Hash0      uint32
	buckets    unsafe.Pointer
	oldbuckets unsafe.Pointer
	nevacuate  uintptr
	extra      *mapextra
}

func GetHmap(t map[int]int) *Hmap {
	if t == nil {
		return nil
	}
	return *(**Hmap)(unsafe.Pointer(&t))
}

func PrintHmap2(t map[int]int, shell *ishell.Shell) {
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
			// На Windows используем стандартный Cyan (36) или Blue (34),
			// которые входят в диапазон 30-37 и не вызывают панику.
			colored = fmt.Sprintf("\x1b[36m%s\x1b[0m", line)
		} else {
			// На Linux/macOS оставляем красивый RGB градиент
			r := start[0] + (end[0]-start[0])*i/steps
			g := start[1] + (end[1]-start[1])*i/steps
			b := start[2] + (end[2]-start[2])*i/steps
			colored = fmt.Sprintf("\x1b[38;2;%d;%d;%dm%s\x1b[0m", r, g, b, line)
		}

		shell.Println(colored)
	}
}

type NoMenuCompleter struct {
	root *ishell.Cmd
}

func (m *NoMenuCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	typed := string(line[:pos])

	// 1. ГЛАВНОЕ: Если строка пустая или только пробелы — возвращаем nil.
	// Это отключает то самое интерактивное меню при нажатии Tab.
	if strings.TrimSpace(typed) == "" {
		return nil, 0
	}

	args := strings.Fields(typed)
	// Проверяем, закончено ли последнее слово пробелом
	isTrailingSpace := len(typed) > 0 && typed[len(typed)-1] == ' '

	// 2. Ищем текущую команду в дереве (используем твой FindCmd)
	currentCmd, _ := m.root.FindCmd(args)
	if currentCmd == nil {
		currentCmd = m.root
	}

	var suggestions [][]rune
	var lastWord string

	if isTrailingSpace {
		// Предлагаем подкоманды для текущей команды
		for _, child := range currentCmd.Children() {
			suggestions = append(suggestions, []rune(child.Name))
		}
		lastWord = ""
	} else {
		// Дописываем текущее слово
		if len(args) > 0 {
			lastWord = args[len(args)-1]
			for _, child := range currentCmd.Children() {
				if strings.HasPrefix(child.Name, lastWord) {
					// Возвращаем только недостающую часть (суффикс)
					suffix := strings.TrimPrefix(child.Name, lastWord)
					suggestions = append(suggestions, []rune(suffix))
				}
			}
		}
	}

	return suggestions, len(lastWord)
}

func main() {
	fmt.Println("131")
	shell := ishell.New()
	shell.CustomCompleter(&NoMenuCompleter{shell.RootCmd()})

	go func() {
		counter := 1
		for {
			time.Sleep(3 * time.Second)

			shell.PrintlnLog(fmt.Sprintf("\n[LOG] Фоновое событие #%d\r", counter))
			// shell.Wait()
			counter++
		}
	}()

	shell.AddCmd(&ishell.Cmd{
		Name: "ping",
		Help: "проверка работы",
		Func: func(c *ishell.Context) {
			c.Println("pong!")
		},
	})

	shell.AddCmd(&ishell.Cmd{
		Name: "pig",
		Help: "проверка работы",
		Func: func(c *ishell.Context) {
			c.Println("onk onk!")
		},
	})
	m := make(map[int]int, 100)

	shell.AddCmd(&ishell.Cmd{
		Name: "hmap",
		Help: "print hmap",
		Func: func(c *ishell.Context) {
			PrintHmap2(m, shell)
		},
	})
	// 3. Запускаем shell
	shell.Run()
}
