package tmpl

import (
	"bytes"
	"io/ioutil"
	"testing"
)

type templatePassCase struct {
	template string
	context  interface{}
	expect   string
}

type templateFailCase struct {
	template string
	context  interface{}
}

type s struct{}

func (s *s) String() string {
	return "foo"
}

func TestTemplateNoContext(t *testing.T) {
	executeTemplatePasses(t, []templatePassCase{
		{`this is just a literal`, nil, `this is just a literal`},
		{`{% block foo %}test{% end block %}{% evoke foo %}`, nil, `test`},
		{`{# foo #}test`, nil, `test`},
		{`{# foo #}test{# bar baz #}`, nil, `test`},
	})
}

func TestTemplatePassBlocks(t *testing.T) {
	executeTemplatePasses(t, []templatePassCase{
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
	})
}

func TestTemplateFailEvoke(t *testing.T) {
	executeTemplateFails(t, []templateFailCase{
		{`{% evoke foo %}`, nil},
	})
}

func TestTemplatePassComments(t *testing.T) {
	executeTemplatePasses(t, []templatePassCase{
		{`{##}`, nil, ``},
		{`{###}`, nil, ``},
		{`{#{#}`, nil, ``},
		{`{#}#}`, nil, ``},
		{"{#\r#}", nil, ``},
		{`{#}foo{#}`, nil, ``},
	})
}

func TestTemplatePassRanges(t *testing.T) {
	executeTemplatePasses(t, []templatePassCase{
		// Space check
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
		// Usage
		{
			`{% range . %}{% .key %}{% .val %}{% end range %}`,
			[]int{0, 1, 2},
			`001122`,
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
		{
			`{% range . as foo bar %}{% .foo %}{% .bar %}{% end range %}`,
			struct{ Foo, Baz string }{"bar", "bif"},
			`FoobarBazbif`,
		},
	})
}

func TestTemplatePassSelections(t *testing.T) {
	executeTemplatePasses(t, []templatePassCase{
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
		{`{% .foo.bar %}`, d{"foo": d{"bar": "baz"}}, `baz`},
		{`{% /.foo.bar %}`, d{"foo": d{"bar": "baz"}}, `baz`},
	})
}

func TestTemplatePassStringers(t *testing.T) {
	type snes struct{ Bar *s }
	executeTemplatePasses(t, []templatePassCase{
		{`{%.%}`, s{}, `{}`},
		{`{%.%}`, &s{}, `foo`},
		{`{%.foo%}`, d{"foo": &s{}}, `foo`},
		{`{%.foo.Bar%}`, d{"foo": &snes{&s{}}}, `foo`},
	})
}

func TestTemplatePassWiths(t *testing.T) {
	executeTemplatePasses(t, []templatePassCase{
		// Space check
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
		// Usage
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
	})
}

func TestTemplateFails(t *testing.T) {
	executeTemplateFails(t, []templateFailCase{
		{`{% $.foo %}`, d{"foo": "bar"}},
		{`{% with . %}{% $.foo %}{% end with %}`, d{"foo": "bar"}},
		{`{% with .foo %}{% $$.foo %}{% end with %}`, d{"foo": "bar"}},
	})
}

func executeTemplateFails(t *testing.T, cases []templateFailCase) {
	for id, c := range cases {
		tree, err := parse(lex([]byte(c.template)))
		if err != nil {
			// If this fires, move to TestLexExpectedFailures in lex_test.go
			// or TestParseExpectedFailures in parse_test.go
			t.Errorf("%d: Parser error: %v", id, err)
			continue
		}
		if err := tree.Execute(ioutil.Discard, c.context); err == nil {
			t.Errorf("%d: Did not fail: %v", id, c.template)
		}
	}
}

func executeTemplatePasses(t *testing.T, cases []templatePassCase) {
	for id, c := range cases {
		tree, err := parse(lex([]byte(c.template)))
		if err != nil {
			t.Errorf("%d: %v", id, err)
			continue
		}
		var buf bytes.Buffer
		if err := tree.Execute(&buf, c.context); err != nil {
			t.Errorf("%d: %v", id, err)
			continue
		}
		if g := buf.String(); g != c.expect {
			t.Errorf("%d\nGot %q\nExp %q", id, g, c.expect)
		}
	}
}
