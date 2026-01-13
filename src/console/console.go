package console

import (
	"fmt"
	"strings"
	"strconv"
	"sync"
	"log"
	"time"

	"visualizer/src/engine"
	"visualizer/src/logger"
	"visualizer/src/cmd"
	"visualizer/src/ws"
	evil "visualizer/collision_mode"

	"github.com/fatih/color"
)

var (
	once sync.Once 

	red    = color.New(color.FgRed)
	yellow = color.New(color.FgYellow)
	green  = color.New(color.FgGreen)
)

func StartConsole[K comparable, V any](t *engine.Type[K, V]) {
	if cmd.Flag.Spectator && !cmd.Flag.Evil {
		spectatorMode[K, V](t)
		return
	}
	
	prob := cmd.Flag.SpectatorFrom

	time.Sleep(time.Millisecond * 150)
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

		cmd0 := strings.ToLower(args[0])

		switch cmd0 {
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
		case "range": 
			//var zerok K 
			//var zerov V 
			if _, ok_k := any(int(0)).(K); !ok_k {
				yellow.Println("Usage(work only with map[int]int): range <int> <int>")
				continue
			}
			if _, ok_v := any(int(0)).(V); !ok_v {
				yellow.Println("Usage(work only with map[int]int): range <int> <int>")
				continue
			}
			if len(args) < 3 {
				yellow.Println("Usage(work only with map[int]int): range <int> <int>")
				continue
			}
			arg_from, _ := strconv.Atoi(args[1])
			arg_to, _ := strconv.Atoi(args[2])

			if arg_from > arg_to {
				yellow.Println("range FROM must be more then range TO")
				continue
			}

			println(arg_from, arg_to)
			
			//currentK = any(arg_from).(K)
			//currentV = any(arg_from).(V)
			//i := arg_from
			update_counter := 0 
			added_counter := 0 
			
			for i := arg_from; i < arg_to; i++ { 
			    currentK := any(i).(K)
			    currentV := any(i).(V)
			    
			    if val, ok := t.Data[currentK]; !ok {
			        added_counter++
			        t.Data[currentK] = currentV
			    } else {
			        if any(val).(int) == any(currentV).(int) {
			            continue 
			        }
			        update_counter++
			        t.Data[currentK] = currentV
			    }
			}

			green.Printf("Range insert successfully: update %d, add %d\n", update_counter, added_counter)
			if update_counter == 0 && added_counter == 0 {continue}
			ws.NotifyUpdate()

			
		case "insert":
			if  cmd.Flag.Evil {
				next := true
				once.Do(func () {
					if _, ok_k := any(int(0)).(K); !ok_k {
						red.Println("Evil mode works only with map[int]int!")
						next = false
					} 					
					if _, ok_v := any(int(0)).(V); !ok_v {
						red.Println("Evil mode works only with map[int]int!")
						next = false
					}
				})

				if !next {continue}
				
				const BUCKET_NUM = uint8(0)
				key := any(prob).(K)
				val := any(prob).(V)
				attempts := 0 
				
				for {
					attempts++

					if BUCKET_NUM == evil.CheckHash(t, key) {
						start := time.Now()
						t.Data[key] = val
						end := time.Since(start)
						yellow.Println("insertion time:", end)
						prob++

						if cmd.Flag.Spectator {
							if prob > cmd.Flag.SpectatorTo {break}
							key = any(prob).(K)
							val = any(prob).(V)
							ws.NotifyUpdate()
							green.Println("Inserted element successfully, attempts:", attempts)
							attempts = 0
							time.Sleep(time.Millisecond * time.Duration(cmd.Flag.Latency))
							continue
						} else {
							break
						}
					}
					prob++
					key = any(prob).(K)
					val = any(prob).(V)
				}

				ws.NotifyUpdate()
				green.Println("Inserted element successfully, attempts:", attempts)
				continue
			} 
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
			red.Println("Unknown command:", cmd0)
		}
	}
}

func evilMode[K comparable, V any](t *engine.Type[K, V]) {

}

func spectatorMode[K comparable, V any](t *engine.Type[K, V]) {
	var zerok K 
	var zerov V
	k := cmd.Flag.SpectatorFrom
	v := cmd.Flag.SpectatorTo

	if k > v {
		log.Println("FROM must be more then TO")
		return
	}

	if val, ok := any(k).(K); ok {
		zerok = val
	} else {
		log.Println("Error", "Bad key type.")
		return
	}
	if val, ok := any(v).(V); ok {
		zerov = val
	} else {
		log.Println("Error", "Bad value type.")
		return
	}

	for {
		if k == cmd.Flag.SpectatorTo {
			break
		}
		t.Data[zerok] = zerov 
		ws.NotifyUpdate()
		time.Sleep(time.Duration(cmd.Flag.Latency) * time.Millisecond)
		zerok = any(k + 1).(K)
		zerov = any(v + 1).(V) 	
		k++ 
		v++
	}
	time.Sleep(time.Second * 2)
	return
}
