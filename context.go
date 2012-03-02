package tmpl

type context struct {
	stack  []interface{}
	blocks map[string]executer
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
