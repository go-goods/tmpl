package tmpl

import (
	"bytes"
	"fmt"
	"strings"
)

type context struct {
	stack  []interface{}
	blocks map[string]executer
	vars   map[string]interface{}
}

func newContext() *context {
	return &context{
		stack:  []interface{}{},
		blocks: map[string]executer{},
		vars:   map[string]interface{}{},
	}
}

func (c *context) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "blocks {")
	for ident, block := range c.blocks {
		fmt.Fprintf(&buf, "\n\t%s: %s", ident, strings.Replace(block.String(), "\n", "\n\t", -1))
	}
	if len(c.blocks) > 0 {
		fmt.Fprint(&buf, "\n")
	}
	fmt.Fprintln(&buf, "}")
	return buf.String()
}

func (c *context) GetBlock(name string) executer {
	return c.blocks[name]
}

func (c *context) Push(val interface{}) {
	//push the value on to the stack and update the value pointer
	c.stack = append(c.stack, val)
}

//grabs the current value from the stack
func (c *context) Current() interface{} {
	return c.stack[len(c.stack)-1]
}

func (c *context) Pop() {
	//slice off the last value
	c.stack = c.stack[:len(c.stack)-1]
}

//perhaps make these just internal details rather than methods
//set and get for vars
func (c *context) Set(key string, val interface{}) {
	c.vars[key] = val
}

func (c *context) Get(key string) interface{} {
	return c.vars[key]
}

func (c *context) Unset(key string) {
	delete(c.vars, key)
}

func (c *context) Exists(key string) (ex bool) {
	_, ex = c.vars[key]
	return
}
