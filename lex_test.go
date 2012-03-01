package tmpl

import "testing"

func TestWorks(t *testing.T) {
	const code = `
		thing
		{% block foo_bar %}
			here
		{% end block %}
		{% call foo .fab "butt" 123.35 -32 %}
		bar
		{% .foo$bar %}
		{% with "val" %}
			{% . %}
		{% end with %}
	`
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
