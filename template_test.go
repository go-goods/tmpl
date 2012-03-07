package tmpl

import (
	"bytes"
	"testing"
)

func TestTemplateExecute(t *testing.T) {
	cases := []struct {
		templ  string
		ctx    interface{}
		expect string
	}{
		// Comments
		{`{##}`, nil, ``},
		{`{###}`, nil, ``},
		{`{#{#}`, nil, ``},
		{`{#}#}`, nil, `#}`},      // ! - Originally expected ``
		{`{#}`, nil, ``},          // !
		{"{#\r#}", nil, ``},       // !
		{`{#}foo{#}`, nil, `foo`}, // !
		// Single-level context selection
		{`{% .foo %}`, d{"foo": "bar"}, `bar`},
		{`{%.foo %}`, d{"foo": "bar"}, `bar`},
		/* FIXME: Expected a "push", got a "0:6[error]invalid character: '%'(MISSING)"
		{`{%.foo%}`, d{"foo": "bar"}, `bar`},
		*/
		/* FIXME: Expected a "push", got a "0:7[error]invalid character: '%'(MISSING)"
		{`{% .foo%}`, d{"foo": "bar"}, `bar`},
		*/
		{`{% /.foo %}`, d{"foo": "bar"}, `bar`},
		{`{% .foo %}`, d{"foo": d{"bar": "baz"}}, `map[bar:baz]`},
		{`{% .foo %}`, d{"foo": 0xBEEF}, `48879`},
		/* FIXME: Got "[98 97 114]"; Exp "bar"
		{`{% .foo %}`, d{"foo": []byte("bar")}, `bar`},
		*/
		// I don't disagree with this output (next 3)
		{`{% .foo %}`, d{"foo": []int{1, 2, 3}}, `[1 2 3]`},
		{`{% .foo %}`, d{"foo": []float64{1, 2, 3}}, `[1 2 3]`},
		{
			`{% .foo %}`,
			d{"foo": []float64{1.41421356, 2.71828183, 3.14159265}},
			`[1.41421356 2.71828183 3.14159265]`,
		},
		// Multi-level context selection
		{`{% .foo.bar %}`, d{"foo": d{"bar": "baz"}}, `baz`},
		{`{% /.foo.bar %}`, d{"foo": d{"bar": "baz"}}, `baz`},
		// Blocks
		{`{% block foo %}{% end block %}`, nil, ``},
		{`{%block foo %}{%end block %}`, nil, ``},
		{`{% block foo%}{% end block%}`, nil, ``},
		/* FIXME: panic: runtime error: invalid memory address or nil pointer dereference [recovered]
		{`{% block foo %}{% end block %}{% evoke foo %}`, nil, ``},
		{`{%block foo%}{%end block%}{% evoke foo %}`, nil, ``},
		{`{%block foo%}{%end block%}{%evoke foo%}`, nil, ``},
		*/
		{
			`{% block foo %}{% .foo %}{% .bar %}{% end block %}{% evoke foo %}`,
			d{"foo": "foo", "bar": "bar"},
			`foobar`,
		},
	}
	for _, c := range cases {
		tree, err := parse(lex([]byte(c.templ)))
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := tree.Execute(&buf, c.ctx); err != nil {
			t.Fatal(err)
		}
		if g := buf.String(); g != c.expect {
			t.Fatalf("\nGot %q\nExp %q", g, c.expect)
		}
	}
}
