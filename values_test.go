package tmpl

import (
	"bytes"
	"reflect"
	"testing"
)

func TestValueParseSelector(t *testing.T) {
	cases := []struct {
		name string
		tmpl string
		sel  *selectorValue
	}{
		{`basic`, `{% .foo.bar %}`, &selectorValue{0, false, []string{"foo", "bar"}}},
		{`rooted`, `{% /.foo.bar %}`, &selectorValue{0, true, []string{"foo", "bar"}}},
		{`relative`, `{% $$.foo.bar %}`, &selectorValue{2, false, []string{"foo", "bar"}}},
		{`previous`, `{% $. %}`, &selectorValue{1, false, nil}},
		{`top`, `{% /. %}`, &selectorValue{0, true, nil}},
		{`empty`, `{% . %}`, &selectorValue{0, false, nil}},
	}

	for _, c := range cases {
		tree, err := parse(lex([]byte(c.tmpl)))
		if err != nil {
			t.Errorf("%s: failed to parse: %s", c.name, err)
			continue
		}

		if !reflect.DeepEqual(tree.base, c.sel) {
			t.Errorf("%s: not equal:\n%v\n%v", c.name, tree, c.sel)
		}
	}
}

type callTest struct {
	tmpl string
	name string
	fn   interface{}
	ctx  interface{}
	exp  string
}

func executePassingCallTests(t *testing.T, cases []callTest) {
	for id, c := range cases {
		tree, err := parse(lex([]byte(c.tmpl)))
		if err != nil {
			t.Errorf("%d: error parsing: %s", id, err)
			continue
		}

		tree.context.funcs[c.name] = reflect.ValueOf(c.fn)
		var buf bytes.Buffer
		if err := tree.Execute(&buf, c.ctx); err != nil {
			t.Errorf("%d: error executing: %s", id, err)
			continue
		}

		if got := buf.String(); got != c.exp {
			t.Errorf("%d:\nExp %q\nGot %q", id, c.exp, got)
		}
	}
}

func TestValueCallPasses(t *testing.T) {
	executePassingCallTests(t, []callTest{
		{
			`{% call foo %}`,
			`foo`,
			func() string { return "foo" },
			nil,
			"foo",
		},
		{
			`{% call foo .foo %}`,
			`foo`,
			func(x string) string { return x },
			d{"foo": "foo"},
			"foo",
		},
		{
			`{% range call foo as _ foo %}{% .foo %}{% end range %}`,
			`foo`,
			func() (string, string) { return "foo", "bar" },
			nil,
			"foobar",
		},
		{
			`{% call add .a .b %}`,
			`add`,
			func(a, b int) int { return a + b },
			d{"a": 10, "b": 20},
			"30",
		},
	})
}

func TestValueBadSelectors(t *testing.T) {
	cases := []struct {
		name string
		tmpl string
	}{
		{`pop inside`, `{% .foo$.bar %}`},
		{`root inside`, `{% .foo/.bar %}`},
		{`double push`, `{% .foo..bar %}`},
		{`pop after root`, `{% /$.foo %}`},
		{`root after pop`, `{% $/.foo %}`},
		{`empty root`, `{% / %}`},
		{`empty pop`, `{% $ %}`},
	}

	for _, c := range cases {
		_, err := parse(lex([]byte(c.tmpl)))
		if err == nil {
			t.Errorf("%s: no error", c.name)
		}
	}
}
