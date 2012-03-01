package tmpl

import "testing"

func TestWorks(t *testing.T) {
	const code = `thing {% block foobar %} here {% end block %} {% call foo .fab "butt" 123.35 -32 %} bar {% .foo$bar %}{% with "val" %}{% . %}{% end with %}`
	for token := range lex([]byte(code)) {
		t.Log(token)
	}
}
