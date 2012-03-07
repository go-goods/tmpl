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
		{`{% with . %}{% end with %}`, nil, ``},
		{`{% block foo %}{% end block %}{% evoke foo %}`, nil, ``},
		{`{%block foo%}{%end block%}{% evoke foo %}`, nil, ``},
		{`{%block foo%}{%end block%}{%evoke foo%}`, nil, ``},
		{
			`{% block foo %}{% .foo %}{% .bar %}{% end block %}{% evoke foo %}`,
			d{"foo": "foo", "bar": "bar"},
			`foobar`,
		},
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
