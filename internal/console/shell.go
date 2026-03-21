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

func (c *Console) SetPrompt(prompt string) {
	c.shell.SetPrompt(prompt)
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

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
)

func (c *Console) PrintlnLogError(a ...any) { c.PrintlnLog(append([]any{red("✗")}, a...)...) }
func (c *Console) PrintlnLogWarn(a ...any)  { c.PrintlnLog(append([]any{yellow("⚠")}, a...)...) }
func (c *Console) PrintlnLogGood(a ...any)  { c.PrintlnLog(append([]any{green("✓")}, a...)...) }

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
	isTrailingSpace := typed[len(typed)-1] == ' '

	var suggestions [][]rune
	var lastWord string

	if isTrailingSpace {
		currentCmd, remaining := m.root.FindCmd(args)
		if currentCmd == nil || len(remaining) > 0 {
			return nil, 0
		}
		for _, child := range currentCmd.Children() {
			suggestions = append(suggestions, []rune(child.Name))
		}
	} else {
		lastWord = args[len(args)-1]
		prefixArgs := args[:len(args)-1]

		var currentCmd *ishell.Cmd
		if len(prefixArgs) == 0 {
			currentCmd = m.root
		} else {
			var remaining []string
			currentCmd, remaining = m.root.FindCmd(prefixArgs)
			if currentCmd == nil || len(remaining) > 0 {
				return nil, 0
			}
		}

		for _, child := range currentCmd.Children() {
			if strings.HasPrefix(child.Name, lastWord) {
				suffix := strings.TrimPrefix(child.Name, lastWord)
				suggestions = append(suggestions, []rune(suffix))
			}
		}
	}

	return suggestions, len(lastWord)
}
