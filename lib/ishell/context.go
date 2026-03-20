package ishell

import (
	"fmt"

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

func logArgs(prefix interface{}, a []interface{}) []interface{} {
	out := make([]interface{}, len(a)+1)
	out[0] = prefix
	copy(out[1:], a)
	return out
}

func (c *Context) PrintlnLogError(a ...interface{}) { c.PrintlnLog(logArgs(red("✗"), a)...) }
func (c *Context) PrintlnLogWarn(a ...interface{})  { c.PrintlnLog(logArgs(yellow("⚠"), a)...) }
func (c *Context) PrintlnLogGood(a ...interface{})  { c.PrintlnLog(logArgs(green("✓"), a)...) }
func (c *Context) PrintlnLogEvent(a ...interface{}) { c.PrintlnLog(logArgs(blue("🛈"), a)...) }

func (c *Context) PrintfLogError(prompt string, a ...interface{}) {
	c.PrintlnLog(red("✗"), fmt.Sprintf(prompt, a...))
}
func (c *Context) PrintfLogWarn(prompt string, a ...interface{}) {
	c.PrintlnLog(yellow("⚠"), fmt.Sprintf(prompt, a...))
}
func (c *Context) PrintfLogGood(prompt string, a ...interface{}) {
	c.PrintlnLog(green("✓"), fmt.Sprintf(prompt, a...))
}
func (c *Context) PrintfLogEvent(prompt string, a ...interface{}) {
	c.PrintlnLog(blue("🛈"), fmt.Sprintf(prompt, a...))
}
func (c *Context) PrintfLogWarn_(prompt string, a ...interface{}) {
	c.Print(yellow("⚠ "), fmt.Sprintf(prompt, a...))
}
func (c *Context) PrintfLog(prompt string, a ...interface{}) {
	c.PrintlnLog(fmt.Sprintf(prompt, a...))
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
