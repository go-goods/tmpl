package tmpl

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type pathItem struct {
	val  reflect.Value
	name string
}

type path []pathItem

func (p path) AbsPath() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "/")
	for _, it := range p {
		fmt.Fprintf(&buf, ".%s", it.name)
	}
	return buf.String()
}

func (p path) ItemBehind(num int) (i pathItem, err error) {
	if num < 0 || num >= len(p) {
		err = fmt.Errorf("%q can't pop %d items off", p.AbsPath(), num)
		return
	}
	i = p[len(p)-(num+1)]
	return
}

func (p path) LastValue() reflect.Value {
	return p[len(p)-1].val
}

type context struct {
	stack  path
	blocks map[string]executer
}

func newContext() *context {
	return &context{
		stack:  path{},
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

func (c *context) getBlock(name string) executer {
	return c.blocks[name]
}

func (c *context) restore(p path) {
	c.stack = p
}

func (c *context) access(key interface{}) (v reflect.Value, err error) {
	//just go hog wild
	defer func() {
		if e := recover(); e != nil {
			v = reflect.Value{}
			err = fmt.Errorf("%q.%q: %q", c.stack.AbsPath(), key, e)
		}
		v = reflect.Indirect(v)
	}()

	val := c.stack.LastValue()
	switch val.Kind() {
	case reflect.Map:
		v = val.MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		if skey, ok := key.(string); !ok {
			err = fmt.Errorf("%q.%q: can't access nonstring key on a struct", c.stack.AbsPath(), key)
		} else {
			v = val.FieldByName(skey)
		}
	default:
		err = fmt.Errorf("%q.%q: cant indirect into %q", c.stack.AbsPath(), key, val.Kind())
	}

	return
}

func (c *context) valueAt(s *selectorValue) (interface{}, error) {
	return nil, nil
}
