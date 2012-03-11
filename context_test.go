package tmpl

import (
	"reflect"
	"testing"
)

type d map[string]interface{}

func TestContextSetPath(t *testing.T) {
	c := newContext()
	c.stack = pathRootedAt(nil)
	sel := &selectorValue{0, false, []string{"foo"}}
	c.set["/.foo"] = reflect.ValueOf("baz")
	val, err := c.valueFor(sel)
	if err != nil {
		t.Fatal(err)
	}
	if val.Interface().(string) != "baz" {
		t.Fatal("Wrong value for selector")
	}
}

func TestContextNestedMaps(t *testing.T) {
	nested := d{"foo": d{"bar": d{"baz": "baz"}}}
	c := newContext()
	c.stack = pathRootedAt(nested)
	sel := &selectorValue{0, false, []string{"foo", "bar", "baz"}}
	val, err := c.valueFor(sel)
	if err != nil {
		t.Fatal(err)
	}
	if val.Interface().(string) != "baz" {
		t.Fatal("Wrong value for selector")
	}
}

func TestContextNestedStructs(t *testing.T) {
	type Baz string
	type Bar struct{ Baz Baz }
	type Foo struct{ Bar Bar }
	type Item struct{ Foo Foo }

	nested := Item{Foo{Bar{Baz("baz")}}}
	c := newContext()
	c.stack = pathRootedAt(nested)
	sel := &selectorValue{0, false, []string{"Foo", "Bar", "Baz"}}
	val, err := c.valueFor(sel)
	if err != nil {
		t.Fatal(err)
	}
	if val.Interface().(Baz) != "baz" {
		t.Fatal("Wrong value for selector")
	}
}

func TestContextInvalidStructKey(t *testing.T) {
	type Baz string
	type Bar struct{ Baz Baz }
	type Foo struct{ Bar Bar }
	type Item struct{ Foo Foo }

	nested := Item{Foo{Bar{Baz("baz")}}}
	c := newContext()
	c.stack = pathRootedAt(nested)
	sel := &selectorValue{0, false, []string{"Foo", "Bar", "az"}}
	_, err := c.valueFor(sel)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestContextInvalidMapKey(t *testing.T) {
	nested := d{"foo": d{"bar": d{"baz": "baz"}}}
	c := newContext()
	c.stack = pathRootedAt(nested)
	sel := &selectorValue{0, false, []string{"foo", "bar", "az"}}
	_, err := c.valueFor(sel)
	if err == nil {
		t.Fatal("expected error")
	}
}
