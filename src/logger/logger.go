package logger

import (
	"fmt"
	"os"
	"sync"

	"github.com/chzyer/readline"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	Log = NewConsole()
)

type Console struct {
	RL     *readline.Instance
	Logger zerolog.Logger
	mu     sync.Mutex
}

func NewConsole() *Console {
	z := log.Output(zerolog.ConsoleWriter{Out: os.Stdout})

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          "> ",
		HistoryFile:     "/tmp/.visualizer_history",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		HistoryLimit:    1000,
	})

	rl.Config.AutoComplete = readline.NewPrefixCompleter(
		readline.PcItem("show"),
		readline.PcItem("hmap"),
		readline.PcItem("delete"),
		readline.PcItem("update"),
		readline.PcItem("insert"),
		readline.PcItem("exit"),
	)
	if err != nil {
		panic(err)
	}

	return &Console{
		RL:     rl,
		Logger: z,
	}
}

func (c *Console) Close() {
	c.RL.Close()
}

func (c *Console) Log(level string, msg string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	event := c.Logger.Info()
	switch level {
	case "debug":
		event = c.Logger.Debug()
	case "error":
		event = c.Logger.Error()
	case "warn":
		event = c.Logger.Warn()
	case "fatal":
		event = c.Logger.Fatal()
	}
	event.Msg(msg)

	if c.RL != nil {
		c.RL.Write([]byte("\r"))
	}
}

func (c *Console) Prompt() (string, error) {
	line, err := c.RL.Readline()
	if err != nil {
		if err == readline.ErrInterrupt {
			return "", fmt.Errorf("interrupted")
		}
		return "", err
	}
	return line, nil
}
