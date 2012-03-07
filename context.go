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

func (p path) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "/")
	for _, it := range p {
		fmt.Fprintf(&buf, ".%s", it.name)
	}
	return buf.String()
}

func (p path) itemBehind(num int) (i pathItem, err error) {
	if num < 0 || num >= len(p) {
		err = fmt.Errorf("%q can't pop %d items off", p, num)
		return
	}
	i = p[len(p)-(num+1)]
	return
}

func (p *path) push(i pathItem) {
	*p = append(*p, i)
}

func (p *path) pop(num int) (err error) {
	if num < 0 || num >= len(*p) {
		err = fmt.Errorf("%q cant pop %d items off", p, num)
		return
	}
	*p = (*p)[:len(*p)-(num)]
	return
}

func (p path) lastValue() reflect.Value {
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

func (c *context) valueFor(s *selectorValue) (v interface{}, err error) {
	return
}

func (c *context) getBlock(name string) executer {
	return c.blocks[name]
}

func access(stack path, val reflect.Value, key string) (v reflect.Value, err error) {
	//just go hog wild
	defer func() {
		if e := recover(); e != nil {
			v = reflect.Value{}
			err = fmt.Errorf("%q.%q: %q", stack, key, e)
		}
		v = reflect.Indirect(v)
	}()

	switch val.Kind() {
	case reflect.Map:
		v = val.MapIndex(reflect.ValueOf(key))
	case reflect.Struct:
		v = val.FieldByName(key)
	default:
		err = fmt.Errorf("%q.%q: cant indirect into %q", stack, key, val.Kind())
	}

	return
}
