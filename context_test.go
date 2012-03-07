package tmpl

import (
	"bytes"
	"testing"
)

type d map[string]interface{}

func TestContextNestedMaps(t *testing.T) {
	nested := d{"foo": d{"bar": d{"baz": "baz"}}}
	c := newContext()
	c.stack = pathRootedAt(nested)
	sel := &selectorValue{0, false, []string{"foo", "bar", "baz"}}
	val, err := c.valueFor(sel)
	if err != nil {
		t.Fatal(err)
	}
	if val.(string) != "baz" {
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
	if val.(Baz) != "baz" {
		t.Fatal("Wrong value for selector")
	}
}

func TestContextAndExecuteMaps(t *testing.T) {
	nested := d{"foo": d{"bar": d{"baz": "baz"}}}
	tmpl := `{% .foo.bar.baz %}`
	tree, err := parse(lex([]byte(tmpl)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, nested); err != nil {
		t.Fatal(err)
	}

	if ex, got := "baz", buf.String(); ex != got {
		t.Fatal("Expected %q. Got %q.", ex, got)
	}
}

func TestContextAndExecuteStructs(t *testing.T) {
	type Baz string
	type Bar struct{ Baz Baz }
	type Foo struct{ Bar Bar }
	type Item struct{ Foo Foo }

	nested := Item{Foo{Bar{Baz("baz")}}}
	tmpl := `{% .Foo.Bar.Baz %}`
	tree, err := parse(lex([]byte(tmpl)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, nested); err != nil {
		t.Fatal(err)
	}

	if ex, got := "baz", buf.String(); ex != got {
		t.Fatal("Expected %q. Got %q.", ex, got)
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

func BenchmarkContextSelectorMap(b *testing.B) {
	nested := d{"foo": d{"bar": d{"baz": "baz"}}}
	c := newContext()
	c.stack = pathRootedAt(nested)
	sel := &selectorValue{0, false, []string{"foo", "bar", "baz"}}
	for i := 0; i < b.N; i++ {
		c.valueFor(sel)
	}
}

func BenchmarkContextSelectorStruct(b *testing.B) {
	type Baz string
	type Bar struct{ Baz Baz }
	type Foo struct{ Bar Bar }
	type Item struct{ Foo Foo }

	nested := Item{Foo{Bar{Baz("baz")}}}
	c := newContext()
	c.stack = pathRootedAt(nested)
	sel := &selectorValue{0, false, []string{"Foo", "Bar", "Baz"}}
	for i := 0; i < b.N; i++ {
		c.valueFor(sel)
	}
}
