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

func TestValueCallNoArgs(t *testing.T) {
	tree, err := parse(lex([]byte(`{% call foo %}`)))
	if err != nil {
		t.Fatal(err)
	}
	tree.context.funcs["foo"] = reflect.ValueOf(func() string { return "foo" })

	var buf bytes.Buffer
	if err := tree.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}

	if got := buf.String(); got != "foo" {
		t.Fatalf("Expected %q. Got %q", "foo", got)
	}
}

func TestValueCallOneArg(t *testing.T) {
	tree, err := parse(lex([]byte(`{% call foo .foo %}`)))
	if err != nil {
		t.Fatal(err)
	}
	tree.context.funcs["foo"] = reflect.ValueOf(func(x string) string { return x })

	var buf bytes.Buffer
	if err := tree.Execute(&buf, d{"foo": "foo"}); err != nil {
		t.Fatal(err)
	}

	if got := buf.String(); got != "foo" {
		t.Fatalf("Expected %q. Got %q", "foo", got)
	}
}

func TestValueCallReturnsRange(t *testing.T) {
	tree, err := parse(lex([]byte(`{% range call foo as _ foo %}{% .foo %}{% end range %}`)))
	if err != nil {
		t.Fatal(err)
	}
	tree.context.funcs["foo"] = reflect.ValueOf(func(string) (string, string) { return "foo", "bar" })

	var buf bytes.Buffer
	if err := tree.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}

	if got := buf.String(); got != "foobar" {
		t.Fatalf("Expected %q. Got %q", "foo", got)
	}
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
