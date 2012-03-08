package tmpl

import (
	"bytes"
	"testing"
)

type s struct{}

func (s *s) String() string {
	return "foo"
}

func TestTemplateExecute(t *testing.T) {
	type snes struct{ Bar *s }

	cases := []struct {
		templ  string
		ctx    interface{}
		expect string
	}{
		// Comments
		{`{##}`, nil, ``},
		{`{###}`, nil, ``},
		{`{#{#}`, nil, ``},
		{`{#}#}`, nil, ``},
		{"{#\r#}", nil, ``},
		{`{#}foo{#}`, nil, ``},
		// Single-level context selection
		{`{% .foo %}`, d{"foo": "bar"}, `bar`},
		{`{%.foo %}`, d{"foo": "bar"}, `bar`},
		{`{%.foo%}`, d{"foo": "bar"}, `bar`},
		{`{% .foo%}`, d{"foo": "bar"}, `bar`},
		{`{% /.foo %}`, d{"foo": "bar"}, `bar`},
		{`{% .foo %}`, d{"foo": d{"bar": "baz"}}, `map[bar:baz]`},
		{`{% .foo %}`, d{"foo": 0xBEEF}, `48879`},
		{`{% .foo %}`, d{"foo": []byte("bar")}, `bar`},
		{`{% .foo %}`, d{"foo": []int{1, 2, 3}}, `[1 2 3]`},
		{`{% .foo %}`, d{"foo": []float64{1, 2, 3}}, `[1 2 3]`},
		{
			`{% .foo %}`,
			d{"foo": []float64{1.41421356, 2.71828183, 3.14159265}},
			`[1.41421356 2.71828183 3.14159265]`,
		},
		// Stringer Satisfactories
		{`{%.%}`, s{}, `{}`},
		{`{%.%}`, &s{}, `foo`},
		{`{%.foo%}`, d{"foo": &s{}}, `foo`},
		{`{%.foo.Bar%}`, d{"foo": &snes{&s{}}}, `foo`},
		// Multi-level context selection
		{`{% .foo.bar %}`, d{"foo": d{"bar": "baz"}}, `baz`},
		{`{% /.foo.bar %}`, d{"foo": d{"bar": "baz"}}, `baz`},
		// Blocks
		{`{% block foo %}{% end block %}`, nil, ``},
		{`{%block foo %}{%end block %}`, nil, ``},
		{`{% block foo%}{% end block%}`, nil, ``},
		{`{% block foo %}{% end block %}{% evoke foo %}`, nil, ``},
		{`{%block foo%}{%end block%}{% evoke foo %}`, nil, ``},
		{`{%block foo%}{%end block%}{%evoke foo%}`, nil, ``},
		{
			`{% block foo %}{% .foo %}{% .bar %}{% end block %}{% evoke foo %}`,
			d{"foo": "foo", "bar": "bar"},
			`foobar`,
		},
		// Range - Space check
		{
			`{% range . as k v %}{% .k %}{% .v %}{% end range %}`,
			[]int{0, 1, 2},
			`001122`,
		},
		{
			`{%range . as k v %}{% .k %}{% .v %}{% end range %}`,
			[]int{0, 1, 2},
			`001122`,
		},
		{
			`{% range . as k v%}{% .k %}{% .v %}{% end range %}`,
			[]int{0, 1, 2},
			`001122`,
		},
		{
			`{% range . as k v %}{% .k %}{% .v %}{%end range %}`,
			[]int{0, 1, 2},
			`001122`,
		},
		{
			`{% range . as k v %}{% .k %}{% .v %}{% end range%}`,
			[]int{0, 1, 2},
			`001122`,
		},
		{
			`{% range . as k v %}{% .k %}{% .v %}{%end range%}`,
			[]int{0, 1, 2},
			`001122`,
		},
		{
			`{%range . as k v%}{% .k %}{% .v %}{% end range %}`,
			[]int{0, 1, 2},
			`001122`,
		},
		{
			`{%range . as k v%}{% .k %}{% .v %}{%end range%}`,
			[]int{0, 1, 2},
			`001122`,
		},
		// Reserved words?
		{
			`{% range . as _ a %}{% .a %}{% end range %}`,
			[]int{0, 1, 2},
			`012`,
		},
		{
			`{% range . as _ range %}{% .range %}{% end range %}`,
			[]int{0, 1, 2},
			`012`,
		},
		{
			`{% range . as _ with %}{% .with %}{% end range %}`,
			[]int{0, 1, 2},
			`012`,
		},
		{
			`{% range . as _ block %}{% .block %}{% end range %}`,
			[]int{0, 1, 2},
			`012`,
		},
		{
			`{% range . as _ if %}{% .if %}{% end range %}`,
			[]int{0, 1, 2},
			`012`,
		},
		{
			`{% range . as _ else %}{% .else %}{% end range %}`,
			[]int{0, 1, 2},
			`012`,
		},
		{
			`{% range . as _ end %}{% .end %}{% end range %}`,
			[]int{0, 1, 2},
			`012`,
		},
		// Range
		{
			`{% range . as foo bar %}{% .foo %}{% .bar %}{% end range%}`,
			[]string{"foo", "bar", "baz"},
			`0foo1bar2baz`,
		},
		{
			`{% range . as foo bar %}{% .foo %}{% .bar %}{% end range %}`,
			d{"foo": "bar"},
			`foobar`,
		},
		{
			`{% range . as foo bar %}{% .foo %}{% .bar %}{% end range %}`,
			struct{ Foo, Baz string }{"bar", "bif"},
			`FoobarBazbif`,
		},
		// With - Space check
		{`{% with . %}{% end with %}`, nil, ``},
		{`{%with . %}{% end with %}`, nil, ``},
		{`{% with .%}{% end with %}`, nil, ``},
		{`{%with .%}{% end with %}`, nil, ``},
		{`{% with . %}{%end with %}`, nil, ``},
		{`{% with . %}{% end with%}`, nil, ``},
		{`{% with . %}{%end with%}`, nil, ``},
		{`{%with .%}{%end with%}`, nil, ``},
		{`{%with .%}{%.foo%}{%end with%}`, d{"foo": "bar"}, `bar`},
		{`{%with .foo%}{%.%}{%end with%}`, d{"foo": "bar"}, `bar`},
		// With - Usage
		{
			`{% with .foo %}{% $.baz %}{% . %}{% end with %}`,
			d{"foo": "bar", "baz": "bif"},
			`bifbar`,
		},
		{
			`{% with .foo.bar.baz.bif %}{% . %}{% $.bif %}{% $$.baz.bif %}{% $$$.bar.baz.bif %}{% end with %}`,
			d{"foo": d{"bar": d{"baz": d{"bif": 0}}}},
			`0000`,
		},
		{
			`{% with .foo.bar.baz.bif %}{% /. %}{% /.foo %}{% /.foo.bar.baz.bif %}{% end with %}`,
			d{"foo": d{"bar": d{"baz": d{"bif": 0}}}},
			`map[foo:map[bar:map[baz:map[bif:0]]]]map[bar:map[baz:map[bif:0]]]0`,
		},
	}
	for id, c := range cases {
		tree, err := parse(lex([]byte(c.templ)))
		if err != nil {
			t.Errorf("%d: %v", id, err)
			continue
		}
		var buf bytes.Buffer
		if err := tree.Execute(&buf, c.ctx); err != nil {
			t.Errorf("%d: %v", id, err)
			continue
		}
		if g := buf.String(); g != c.expect {
			t.Errorf("%d\nGot %q\nExp %q", id, g, c.expect)
		}
	}
}
