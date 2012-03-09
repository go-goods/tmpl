package tmpl

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

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

func BenchmarkExecuteProduction(b *testing.B) {
	b.StopTimer()
	dir := createTestDir(b, []templateFile{
		{"base.tmpl", `{% evoke foo %}`},
		{"foo.block", `{% block foo %}some data{% end block%}`},
	})
	defer os.RemoveAll(dir)

	j := func(path string) string {
		return filepath.Join(dir, path)
	}
	tmp := Parse(j("base.tmpl"))
	tmp.Blocks(j("foo.block"))
	tmp.Execute(ioutil.Discard, nil)
	defer CompileMode(<-modeChan)
	CompileMode(Production)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tmp.Execute(ioutil.Discard, nil)
	}
}

func BenchmarkExecuteProductionReload(b *testing.B) {
	b.StopTimer()
	dir := createTestDir(b, []templateFile{
		{"base.tmpl", `{% evoke foo %}{% evoke bar %}`},
		{"foo.block", `{% block foo %}some data{% end block%}`},
		{"bar.block", `{% block bar %}some more data{% end block%}`},
	})
	defer os.RemoveAll(dir)

	j := func(path string) string {
		return filepath.Join(dir, path)
	}
	tmp := Parse(j("base.tmpl"))
	tmp.Blocks(j("foo.block"))
	bar := j("bar.block")
	tmp.Execute(ioutil.Discard, nil)
	defer CompileMode(<-modeChan)
	CompileMode(Production)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tmp.Execute(ioutil.Discard, bar)
	}
}

func BenchmarkExecuteDevelopment(b *testing.B) {
	b.StopTimer()
	dir := createTestDir(b, []templateFile{
		{"base.tmpl", `{% evoke foo %}`},
		{"foo.block", `{% block foo %}some data{% end block%}`},
	})
	defer os.RemoveAll(dir)

	j := func(path string) string {
		return filepath.Join(dir, path)
	}
	tmp := Parse(j("base.tmpl"))
	tmp.Blocks(j("foo.block"))
	tmp.Execute(ioutil.Discard, nil)
	defer CompileMode(<-modeChan)
	CompileMode(Development)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tmp.Execute(ioutil.Discard, nil)
	}
}

func BenchmarkExecuteDevelopmentReload(b *testing.B) {
	b.StopTimer()
	dir := createTestDir(b, []templateFile{
		{"base.tmpl", `{% evoke foo %}{% evoke bar %}`},
		{"foo.block", `{% block foo %}some data{% end block%}`},
		{"bar.block", `{% block bar %}some more data{% end block%}`},
	})
	defer os.RemoveAll(dir)

	j := func(path string) string {
		return filepath.Join(dir, path)
	}
	tmp := Parse(j("base.tmpl"))
	tmp.Blocks(j("foo.block"))
	bar := j("bar.block")
	tmp.Execute(ioutil.Discard, nil)
	defer CompileMode(<-modeChan)
	CompileMode(Development)

	b.StartTimer()
	for i := 0; i < b.N; i++ {
		tmp.Execute(ioutil.Discard, bar)
	}
}
