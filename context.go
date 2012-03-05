package tmpl

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type context struct {
	stack  []reflect.Value
	blocks map[string]executer
}

func newContext() *context {
	return &context{
		stack:  []reflect.Value{},
		blocks: map[string]executer{},
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
	c.stack = append(c.stack, reflect.ValueOf(val))
}

func (c *context) ContextAt(pops int) (reflect.Value, error) {
	if pops < 0 {
		return reflect.Value{}, fmt.Errorf("negative number of pops")
	}
	if pops >= len(c.stack) {
		return reflect.Value{}, fmt.Errorf("too many pops")
	}
	return c.stack[len(c.stack)-(pops+1)], nil
}

func (c *context) Pop() {
	//slice off the last value
	c.stack = c.stack[:len(c.stack)-1]
}
