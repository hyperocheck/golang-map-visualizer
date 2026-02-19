package ishell

import (
	"github.com/fatih/color"
)

// Context is an ishell context. It embeds ishell.Actions.
type Context struct {
	contextValues
	err error

	// Args is command arguments.
	Args []string

	// RawArgs is unprocessed command arguments.
	RawArgs []string

	// Cmd is the currently executing command. This is empty for NotFound and Interrupt.
	Cmd Cmd

	Actions
}

// Err informs ishell that an error occurred in the current
// function.
func (c *Context) Err(err error) {
	c.err = err
}

var (
	red    = color.New(color.FgRed).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
)

func (c *Context) PrintlnLogError(a ...interface{}) {
	logs := make([]interface{}, 0, len(a)+1)
	logs = append(logs, red("ERR"))
	logs = append(logs, a...)

	c.PrintlnLog(logs...)
}

func (c *Context) PrintlnLogWarn(a ...interface{}) {
	logs := make([]interface{}, 0, len(a)+1)
	logs = append(logs, yellow("WARN"))
	logs = append(logs, a...)

	c.PrintlnLog(logs...)
}

func (c *Context) PrintlnLogGood(a ...interface{}) {
	logs := make([]interface{}, 0, len(a)+1)
	logs = append(logs, green("GOOD"))
	logs = append(logs, a...)

	c.PrintlnLog(logs...)
}

func (c *Context) PrintlnLogEvent(a ...interface{}) {
	logs := make([]interface{}, 0, len(a)+1)
	logs = append(logs, blue("INFO"))
	logs = append(logs, a...)

	c.PrintlnLog(logs...)
}

func (c *Context) PrintfLogError(format string, a ...interface{}) {
	format = red("ERR ") + format

	c.Printf(format, a...)
}

func (c *Context) PrintfLogWarn(format string, a ...interface{}) {
	format = yellow("WARN ") + format

	c.Printf(format, a...)
}

func (c *Context) PrintfLogGood(format string, a ...interface{}) {
	format = green("GOOD ") + format

	c.Printf(format, a...)
}

func (c *Context) PrintfLogEvent(format string, a ...interface{}) {
	format = blue("INFO ") + format

	c.Printf(format, a...)
}

func (c *Context) PrintfLog(format string, a ...interface{}) {
	c.Printf(format, a...)
}

// contextValues is the map for values in the context.
type contextValues map[string]interface{}

// Get returns the value associated with this context for key, or nil
// if no value is associated with key. Successive calls to Get with
// the same key returns the same result.
func (c contextValues) Get(key string) interface{} {
	return c[key]
}

// Set sets the key in this context to value.
func (c *contextValues) Set(key string, value interface{}) {
	if *c == nil {
		*c = make(map[string]interface{})
	}
	(*c)[key] = value
}

// Del deletes key and its value in this context.
func (c contextValues) Del(key string) {
	delete(c, key)
}

// Keys returns all keys in the context.
func (c contextValues) Keys() (keys []string) {
	for key := range c {
		keys = append(keys, key)
	}
	return
}
