package tmpl

import "testing"

type parseFailCase struct {
	code string
}

func TestParseExpectedFailures(t *testing.T) {
	executeParseFails(t, []parseFailCase{
		{`{%%}`},
		{`{% %}`},
		{`{% block %}`},
		{`{% block foo %}`},
		{`{% block %}{% end block %}`},
		{`{% else %}`},
		{`{% if %}`},
		{`{% if . %}`},
		{`{% if %}{% end if %}`},
		{`{% range %}`},
		{`{% range .foo %}`},
		{`{% range .foo as %}`},
		{`{% range .foo as bar %}`},
		{`{% range .foo as bar baz %}`},
		{`{% range %}{% end range %}`},
		{`{% range .foo as %}{% end range %}`},
		{`{% range .foo as bar %}{% end range %}`},
		{`{% with %}`},
		{`{% with . %}`},
		{`{% with %}{% end with %}`},
		{`{% end %}`},
		{`{% end block %}`},
		{`{% end if %}`},
		{`{% end range %}`},
		{`{% end with %}`},
		{`{% block foo %}{% block bar %}{% end block %}{% end block %}`},
	})
}

func executeParseFails(t *testing.T, cases []parseFailCase) {
	for id, c := range cases {
		for token := range lex([]byte(c.code)) {
			if token.typ == tokenError {
				t.Errorf("%d: Should lex: %s", id, c.code)
				continue
			}
		}
		_, err := parse(lex([]byte(c.code)))
		if err == nil {
			t.Errorf("%d: Should not parse: %s", id, c.code)
		}
	}
}
