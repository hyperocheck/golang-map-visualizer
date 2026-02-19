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

func (c *Context) PrintlnLogError(a ...interface{}) {
	red := color.New(color.FgRed).SprintFunc()
	c.PrintlnLog(red("[ERR]"), a)
}

func (c *Context) PrintlnLogWarn(a ...interface{}) {
	yellow := color.New(color.FgYellow).SprintFunc()
	c.PrintlnLog(yellow("[WARN]"), a)
}

func (c *Context) PrintlnLogGood(a ...interface{}) {
	green := color.New(color.FgGreen).SprintFunc()
	c.PrintlnLog(green("[GOOD]"), a)
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
