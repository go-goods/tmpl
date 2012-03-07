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
