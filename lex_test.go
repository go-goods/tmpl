package tmpl

import "testing"

func TestLexEspected(t *testing.T) {
	cases := []struct {
		name string
		code string
		ex   []tokenType
	}{
		{`pop`, `{% $foo %}`, []tokenType{tokenOpen, tokenStartSel, tokenPop, tokenIdent, tokenEndSel, tokenClose, tokenEOF}},
		{`comment`, `{# with .foo #}`, []tokenType{tokenComment, tokenEOF}},
		{`comment suffocate`, `{#No spaces#}`, []tokenType{tokenComment, tokenEOF}},
		{`comment nl`, `{# My comment spans
			two lines #}`, []tokenType{tokenComment, tokenEOF}},
		{`complicated selectors`, `{% call func .foo.bar $$.bar.baz.foo /.foo.bar %}`,
			[]tokenType{tokenOpen, tokenCall, tokenIdent, tokenStartSel, tokenPush, tokenIdent,
				tokenPush, tokenIdent, tokenEndSel, tokenStartSel, tokenPop, tokenPop, tokenPush,
				tokenIdent, tokenPush, tokenIdent, tokenPush, tokenIdent, tokenEndSel, tokenStartSel,
				tokenRoot, tokenPush, tokenIdent, tokenPush, tokenIdent, tokenEndSel, tokenClose, tokenEOF},
		},
		{`as`, `{% range . as _ val %}`, []tokenType{tokenOpen, tokenRange, tokenStartSel, tokenPush, tokenEndSel, tokenAs, tokenIdent, tokenIdent, tokenClose, tokenEOF}},
		{`if`, `{% if .foo %}`, []tokenType{tokenOpen, tokenIf, tokenStartSel, tokenPush, tokenIdent, tokenEndSel, tokenClose, tokenEOF}},
		{`crazy comment`, `{#}#}`, []tokenType{tokenComment, tokenEOF}},
		{`multi crazy`, `{#}foo{#}`, []tokenType{tokenComment, tokenEOF}},
		{`selector suffocate`, `{%.%}`, []tokenType{tokenOpen, tokenStartSel, tokenPush, tokenEndSel, tokenClose, tokenEOF}},
	}

	for _, c := range cases {
		var toks []tokenType
		for token := range lex([]byte(c.code)) {
			toks = append(toks, token.typ)
			if token.typ == tokenError {
				t.Errorf("%s: Unexpected error: %v\n", c.name, token)
			}
		}
		if len(c.ex) != len(toks) {
			t.Errorf("%s: Expected %v got %v\n", c.name, c.ex, toks)
		}
		for i, typ := range c.ex {
			if toks[i] != typ {
				t.Errorf("%s:%d: Expected a %v got a %v\n", c.name, i, typ, toks[i])
			}
		}
	}
}

func TestLexExpectedFailures(t *testing.T) {
	cases := []struct {
		code string
	}{
		{`{#}`},
		{`{%}`},
	}
caseBlock:
	for id, c := range cases {
		for token := range lex([]byte(c.code)) {
			if token.typ == tokenError {
				continue caseBlock
			}
		}
		t.Errorf("%d: Should not lex: %s", id, c.code)
	}
}

func TestLexAllTokensNamed(t *testing.T) {
	if len(tokenNames) != int(tokenError)+1 {
		t.Fatalf("%d tokens %d names", tokenError+1, len(tokenNames))
	}
}
