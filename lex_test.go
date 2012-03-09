package tmpl

import "testing"

func TestLexExpected(t *testing.T) {
	cases := []struct {
		code string
		ex   []tokenType
	}{
		{`{% $foo %}`, []tokenType{tokenOpen, tokenStartSel, tokenPop, tokenIdent, tokenEndSel, tokenClose, tokenEOF}},
		{`{# with .foo #}`, []tokenType{tokenComment, tokenEOF}},
		{`{#No spaces#}`, []tokenType{tokenComment, tokenEOF}},
		{`{# My comment spans
			two lines #}`, []tokenType{tokenComment, tokenEOF}},
		{`{% call func .foo.bar $$.bar.baz.foo /.foo.bar %}`,
			[]tokenType{tokenOpen, tokenCall, tokenIdent, tokenStartSel, tokenPush, tokenIdent,
				tokenPush, tokenIdent, tokenEndSel, tokenStartSel, tokenPop, tokenPop, tokenPush,
				tokenIdent, tokenPush, tokenIdent, tokenPush, tokenIdent, tokenEndSel, tokenStartSel,
				tokenRoot, tokenPush, tokenIdent, tokenPush, tokenIdent, tokenEndSel, tokenClose, tokenEOF},
		},
		{`{% range . as _ val %}`, []tokenType{tokenOpen, tokenRange, tokenStartSel, tokenPush, tokenEndSel, tokenAs, tokenIdent, tokenIdent, tokenClose, tokenEOF}},
		{`{% if .foo %}`, []tokenType{tokenOpen, tokenIf, tokenStartSel, tokenPush, tokenIdent, tokenEndSel, tokenClose, tokenEOF}},
		{`{#}#}`, []tokenType{tokenComment, tokenEOF}},
		{`{#}foo{#}`, []tokenType{tokenComment, tokenEOF}},
		{`{%.%}`, []tokenType{tokenOpen, tokenStartSel, tokenPush, tokenEndSel, tokenClose, tokenEOF}},
		{`{% range .foo as _ rangev %}`, []tokenType{tokenOpen, tokenRange, tokenStartSel, tokenPush, tokenIdent, tokenEndSel, tokenAs, tokenIdent, tokenIdent, tokenClose, tokenEOF}},
		{`{% block block1 %}`, []tokenType{tokenOpen, tokenBlock, tokenIdent, tokenClose, tokenEOF}},
	}

	for id, c := range cases {
		var toks []tokenType
		for token := range lex([]byte(c.code)) {
			toks = append(toks, token.typ)
			if token.typ == tokenError {
				t.Errorf("%d: Unexpected error: %v\n", id, token)
			}
		}
		if len(c.ex) != len(toks) {
			t.Errorf("%d: Expected %v got %v\n", id, c.ex, toks)
			continue
		}
		for i, typ := range c.ex {
			if toks[i] != typ {
				t.Errorf("%d: %d: Expected a %v got a %v\n", id, i, typ, toks[i])
			}
		}
	}
}

func TestLexExpectedFailures(t *testing.T) {
	cases := []struct {
		code string
	}{
		{`{#`},
		{`{#}`},
		{`{%`},
		{`{%}`},
		{`{% {% . %}`},
		{`{% {% %} . %}`},
		{`{% {# . %}`},
		{`{% {# #} . %}`},
		{`{% ~ %}`},
		{`{% ! %}`},
		{`{% @ %}`},
		{`{% # %}`},
		{`{% % %}`},
		{`{% ^ %}`},
		{`{% & %}`},
		{`{% * %}`},
		{`{% ( %}`},
		{`{% ) %}`},
		{`{% - %}`},
		{`{% + %}`},
		{`{% = %}`},
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
