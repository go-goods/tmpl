package tmpl

import "testing"

func TestLexEspected(t *testing.T) {
	cases := []struct {
		name string
		code string
		ex   []tokenType
	}{
		{`pop`, `{% $foo %}`, []tokenType{tokenOpen, tokenStartSel, tokenPop, tokenIdent, tokenEndSel, tokenClose, tokenEOF}},
		{`numeric`, `{% with 25 %}`, []tokenType{tokenOpen, tokenWith, tokenNumeric, tokenClose, tokenEOF}},
		{`variable`, `{% with ^foo %}`, []tokenType{tokenOpen, tokenWith, tokenStartSel, tokenVar, tokenIdent, tokenEndSel, tokenClose, tokenEOF}},
		{`comment`, `{# with ^foo #}`, []tokenType{tokenComment, tokenEOF}},
		{`as`, `{% range . as _ val %}`, []tokenType{tokenOpen, tokenRange, tokenStartSel, tokenPush, tokenEndSel, tokenAs, tokenIdent, tokenIdent, tokenClose, tokenEOF}},
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

func TestLexAllTokensNamed(t *testing.T) {
	if len(tokenNames) != int(tokenError)+1 {
		t.Fatalf("%d tokens %d names", tokenError+1, len(tokenNames))
	}
}
