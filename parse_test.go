package tmpl

import "testing"

type parseTestCase struct {
	code string
}

func TestParseExpectedFailures(t *testing.T) {
	executeParseFails(t, []parseTestCase{
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

func TestParseExpectedPasses(t *testing.T) {
	executeParsePasses(t, []parseTestCase{
		{`{% range call foo as _ foo %}{% end range %}`},
	})
}

func executeParsePasses(t *testing.T, cases []parseTestCase) {
	for id, c := range cases {
		for token := range lex([]byte(c.code)) {
			if token.typ == tokenError {
				t.Errorf("%d: Should lex: %s\n%s", id, c.code, token)
				continue
			}
		}
		_, err := parse(lex([]byte(c.code)))
		if err != nil {
			t.Errorf("%d: Should parse: %s\n%s", id, c.code, err)
		}
	}
}

func executeParseFails(t *testing.T, cases []parseTestCase) {
	for id, c := range cases {
		for token := range lex([]byte(c.code)) {
			if token.typ == tokenError {
				t.Errorf("%d: Should lex: %s\n%s", id, c.code, token)
				continue
			}
		}
		_, err := parse(lex([]byte(c.code)))
		if err == nil {
			t.Errorf("%d: Should not parse: %s\n%s", id, c.code, err)
		}
	}
}
