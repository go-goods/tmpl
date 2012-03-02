package tmpl

import "testing"

func TestWorks(t *testing.T) {
	const code = `literal {% call .foo$bar .bar$$baz..foo 2.3 5e7 "boof" %}
	{% block foo %}{% end block %}`
	for token := range lex([]byte(code)) {
		// t.Log(token)
		if token.typ == tokenError {
			t.Error("Unexpected error:", token)
		}
	}
}

func TestSelPopStart(t *testing.T) {
	const code = `{% $foo %}`
	for token := range lex([]byte(code)) {
		if token.typ == tokenError {
			t.Error("Unexpected error:", token)
		}
	}
}

func TestAllTokensNamed(t *testing.T) {
	if len(tokenNames) != int(tokenError)+1 {
		t.Fatalf("%d tokens %d names", tokenError+1, len(tokenNames))
	}
}
