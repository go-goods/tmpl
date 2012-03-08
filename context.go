package tmpl

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

func indirect(v reflect.Value) reflect.Value {
	for {
		if v.Kind() == reflect.Interface && !v.IsNil() {
			v = v.Elem()
			continue
		}
		if v.Kind() != reflect.Ptr || v.IsNil() {
			break
		}
		v = v.Elem()
	}
	return v
}

func access(stack path, val reflect.Value, key string) (v reflect.Value, err error) {
	//just go hog wild
	defer func() {
		if e := recover(); e != nil {
			v = reflect.Value{}
			err = fmt.Errorf("%q.%q: %q", stack, key, e)
		}
	}()

	val = indirect(val)
	switch val.Kind() {
	case reflect.Map:
		v = val.MapIndex(reflect.ValueOf(key))
		if !v.IsValid() {
			err = fmt.Errorf("%q.%q: field not found", stack, key)
		}
	case reflect.Struct:
		v = val.FieldByName(key)
		if !v.IsValid() {
			err = fmt.Errorf("%q.%q: field not found", stack, key)
		}
	default:
		err = fmt.Errorf("%q.%q: cant indirect into %q", stack, key, val.Kind())
	}

	return
}

type context struct {
	stack  path
	blocks map[string]*executeBlockValue
	backup map[string]*executeBlockValue
	funcs  map[string]reflect.Value
	set    map[string]interface{}
}

func newContext() *context {
	return &context{
		stack:  path{},
		funcs:  map[string]reflect.Value{},
		blocks: map[string]*executeBlockValue{},
		set:    map[string]interface{}{},
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

//sets what file the blocks on context was generated from
func (c *context) setFile(file string) {
	for _, val := range c.blocks {
		val.file = file
	}
}

func (c *context) dup() {
	c.backup = map[string]*executeBlockValue{}
	for key := range c.blocks {
		c.backup[key] = c.blocks[key]
	}
}

func (c *context) restore() {
	c.blocks = map[string]*executeBlockValue{}
	for key := range c.backup {
		c.blocks[key] = c.backup[key]
	}
}

func (c *context) valueFor(s *selectorValue) (v interface{}, err error) {
	var pth path
	switch {
	case s == nil:
		err = fmt.Errorf("%q: can't get the value for a nil selector", c.stack)
		return
	case s.abs:
		pth = path(c.stack[:1])
	case s.pops < 0 || s.pops >= len(c.stack):
		err = fmt.Errorf("%q: cant pop %d items", c.stack, s.pops)
		return
	case s.pops > 0:
		pth = path(c.stack[:len(c.stack)-(s.pops)])
	default:
		pth = c.stack
	}

	//check our path override for that value
	if iv, ex := c.set[pth.StringWith(s.path)]; ex {
		v = iv
		return
	}

	rv, err := pth.valueAt(s.path)
	if err != nil {
		return
	}
	v = rv.Interface()
	return
}

func (c *context) cd(s *selectorValue) (err error) {
	switch {
	case s == nil:
		err = fmt.Errorf("%q: can't get the value for a nil selector", c.stack)
		return
	case s.abs:
		c.stack = c.stack[:1]
	case s.pops < 0 || s.pops >= len(c.stack):
		err = fmt.Errorf("%q: cant pop %d items", c.stack, s.pops)
		return
	case s.pops > 0:
		c.stack = c.stack[:len(c.stack)-s.pops]
	}
	err = c.stack.cd(s.path)
	return
}

func (c *context) setStack(p path) {
	c.stack = p
}

func (c *context) getBlock(name string) *executeBlockValue {
	return c.blocks[name]
}

func (c *context) setAt(path string, value interface{}) {
	if path != "" {
		c.set[path] = value
	}
}

func (c *context) unsetAt(path string) {
	if path != "" {
		delete(c.set, path)
	}
}
