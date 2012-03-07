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
		if v.Kind() != reflect.Ptr {
			break
		}
		if v.IsNil() {
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
		v = indirect(v)
	}()

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
	var rv reflect.Value
	switch {
	case s.abs:
		rv, err = path(c.stack[:1]).valueAt(s.path)
		if err != nil {
			return
		}
		v = rv.Interface()
	case s.pops > 0:
		rv, err = path(c.stack[:len(c.stack)-(s.pops)]).valueAt(s.path)
		if err != nil {
			return
		}
		v = rv.Interface()
	default:
		rv, err = c.stack.valueAt(s.path)
		if err != nil {
			return
		}
		v = rv.Interface()
	}
	return
}

func (c *context) setStack(p path) {
	c.stack = p
}

func (c *context) getBlock(name string) executer {
	return c.blocks[name]
}
