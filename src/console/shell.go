package console

import (
	"strings"

	"github.com/abiosoft/ishell/v2"
	"github.com/fatih/color"
)

type Console struct {
	shell *ishell.Shell
}

type CommandHandler func(ctx *ishell.Context)

func New() *Console {
	shell := ishell.New()
	shell.CustomCompleter(&NoMenuCompleter{shell.RootCmd()})
	return &Console{shell: shell}
}

func (c *Console) Print(a ...any) {
	c.shell.Print(a...)
}

func (c *Console) Println(a ...any) {
	c.shell.Println(a...)
}

func (c *Console) PrintlnLog(a ...any) {
	c.shell.PrintlnLog(a...)
}

func (c *Console) Printf(format string, a ...any) {
	c.shell.Printf(format, a...)
}

func (c *Console) PrintlnLogError(a ...any) {
	red := color.New(color.FgRed).SprintFunc()
	c.shell.PrintlnLog(red("[ERR]"), a)
}

func (c *Console) PrintlnLogWarn(a ...any) {
	yellow := color.New(color.FgYellow).SprintFunc()
	c.shell.PrintlnLog(yellow("[WARN]"), a)
}

func (c *Console) PrintlnLogGood(a ...any) {
	green := color.New(color.FgGreen).SprintFunc()
	c.shell.PrintlnLog(green("[GOOD]"), a)
}

func (c *Console) RegisterCommand(name, help string, handler CommandHandler) {
	c.shell.AddCmd(&ishell.Cmd{
		Name: name,
		Help: help,
		Func: handler,
	})
}

func (c *Console) Run() {
	c.shell.Run()
}

type NoMenuCompleter struct {
	root *ishell.Cmd
}

func (m *NoMenuCompleter) Do(line []rune, pos int) (newLine [][]rune, length int) {
	typed := string(line[:pos])

	if strings.TrimSpace(typed) == "" {
		return nil, 0
	}

	args := strings.Fields(typed)
	isTrailingSpace := len(typed) > 0 && typed[len(typed)-1] == ' '

	currentCmd, _ := m.root.FindCmd(args)
	if currentCmd == nil {
		currentCmd = m.root
	}

	var suggestions [][]rune
	var lastWord string

	if isTrailingSpace {
		for _, child := range currentCmd.Children() {
			suggestions = append(suggestions, []rune(child.Name))
		}
		lastWord = ""
	} else {
		if len(args) > 0 {
			lastWord = args[len(args)-1]
			for _, child := range currentCmd.Children() {
				if strings.HasPrefix(child.Name, lastWord) {
					suffix := strings.TrimPrefix(child.Name, lastWord)
					suggestions = append(suggestions, []rune(suffix))
				}
			}
		}
	}

	return suggestions, len(lastWord)
}
