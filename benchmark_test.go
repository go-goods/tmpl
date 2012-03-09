package tmpl

import "testing"

func BenchmarkPathStringWith(b *testing.B) {
	pth := pathRootedAt(nil)
	sel := []string{"foo", "bar", "baz"}

	for i := 0; i < b.N; i++ {
		pth.StringWith(sel)
	}
}

func BenchmarkParseSpeed(b *testing.B) {
	t := []byte(testLongTemplate)
	for i := 0; i < b.N; i++ {
		parse(lex(t))
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
