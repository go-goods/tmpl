package tmpl

import (
	"bytes"
	"testing"
)

func TestExecuteNoContext(t *testing.T) {
	cases := []struct {
		templ  string
		expect string
	}{
		{`this is just a literal`, `this is just a literal`},
		{`{% if 1 %}test{% end if %}`, `test`},
		{`{% if 1 %}test{% else %}fail{% end if %}`, `test`},
		{`{% block foo %}test{% end block %}`, `test`},
		{`t{%%}e{%%}s{%%}t{%%}`, `test`},
	}
	for _, c := range cases {
		tree, err := parse(lex([]byte(c.templ)))
		if err != nil {
			t.Fatal(err)
		}
		var buf bytes.Buffer
		if err := tree.Execute(&buf, nil); err != nil {
			t.Fatal(err)
		}
		if g := buf.String(); g != c.expect {
			t.Fatalf("\nGot %q\nExp %q", g, c.expect)
		}
	}
}
