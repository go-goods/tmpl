package tmpl

import (
	"bytes"
	"fmt"
	"reflect"
)

type pathItem struct {
	val  reflect.Value
	name string
}

type path []pathItem

func pathRootedAt(v interface{}) path {
	return path{pathItem{
		name: "",
		val:  reflect.ValueOf(v),
	}}
}

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

func (p path) dup() (d path) {
	copy(d, p)
	return
}

func (p *path) cd(keys []string) error {
	for _, key := range keys {
		val, err := access(*p, p.lastValue(), key)
		if err != nil {
			return err
		}

		p.push(pathItem{
			name: key,
			val:  val,
		})
	}
	return nil
}

func (p path) valueAt(keys []string) (v reflect.Value, err error) {
	v = p.lastValue()
	for i, key := range keys {
		v, err = access(p, v, key)
		if err != nil {
			return v, fmt.Errorf("%q: Error accessing item %d: %q", p, i, key)
		}
	}
	return
}
