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

func TestContextAndExecute(t *testing.T) {
	tmpl := `{% .foo.bar.baz %}`
	tree, err := parse(lex([]byte(tmpl)))
	if err != nil {
		t.Fatal(err)
	}
	nested := d{"foo": d{"bar": d{"baz": "baz"}}}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, nested); err != nil {
		t.Fatal(err)
	}

	if ex, got := "baz", buf.String(); ex != got {
		t.Fatal("Expected %q. Got %q.", ex, got)
	}
}
