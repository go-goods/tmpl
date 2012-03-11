package tmpl

import (
	"reflect"
	"testing"
)

const testLongTemplate = `
	literal
	{% call func .foo.bar $$.bar.baz.foo /.foo.bar %}
	{% block baz %}
		doof bood
		{% range .bar %}
			ding dong
		{% end range %}
		{% with .bar %}
			{% evoke butt %}
		{% end with %}
	{% end block %}
	{% block butt %}
		dar fangle {% if .foo %}doo{% else %}no doo{% end if %}
	{% end block %}
	{% evoke baz %}
	{% if .foo %}
		always!
	{% end if %}{% if .bar %}{% else %}doof{% end if %}
	{% with .ff %}{% . %}{% end with %}
	{% range . as _ val %}{% with .val %}{% . %}{% end with %}{% end range %}
`

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
		{`{% else if %}`},
		{`{% else if . %}`},
		{`{% if %}`},
		{`{% if . %}`},
		{`{% if %}{% end if %}`},
		{`{% if bad %}{% end if %}`},
		{`{% range %}`},
		{`{% range .foo %}`},
		{`{% range .foo as %}`},
		{`{% range .foo as bar %}`},
		{`{% range .foo as bar baz %}`},
		{`{% range %}{% end range %}`},
		{`{% range .foo as %}{% end range %}`},
		{`{% range .foo as bar %}{% end range %}`},
		{`{% range bad %}{% end range %}`},
		{`{% with %}`},
		{`{% with . %}`},
		{`{% with %}{% end with %}`},
		{`{% with bad %}{% end with %} %}`},
		{`{% end %}`},
		{`{% end block %}`},
		{`{% end if %}`},
		{`{% end range %}`},
		{`{% end with %}`},
		{`{% block foo %}{% block bar %}{% end block %}{% end block %}`},
		{`{% block foo %}{% with . %}{% block bar %}{% end block %}{% end with %}{% end block %}`},
		{`{% block foo %}{% end block %}{% block foo %}{% end block %}`},
		{`{% block .foo %}{% end block %}`},

		//cant evoke or with on things other than selectors
		{`{% evoke foo ctx %}`},
		{`{% evoke call foo %}`},
		{`{% evoke foo call foo %}`},
		{`{% with call foo %}`},
		{`{% with call foo %}{% end with %}`},

		//unknown keyword
		{`{% flabdab %}`},
		{`{% range foo flabdab %}{% end range %}`},
		{`{% block foo flabdab %}{% end block %}`},
		{`{% call foo flabdab %}`},
		{`{% if .foo flabdab %}{% end if %}`},
		{`{% with .foo flabdab %}{% end with %}`},
		{`{% if .foo %}{% else flabdab %}{% end if %}`},
		{`{% elseif %}`},
	})
}

func TestParseExpectedPasses(t *testing.T) {
	executeParsePasses(t, []parseTestCase{
		{`{% range call foo as _ foo %}{% end range %}`},
		{testLongTemplate},
		{``},
		{`{% if . %}{% else %}{% end if %}`},
		{`{% if . %}{% end if %}`},
		{`{% if call foo %}{% end if %}`},
		{`{% if call foo %}{% else %}{% end if %}`},
		{`{% block foo %}{% end block %}`},
		{`{% block foo %}{% with . %}{% end with %}{% end block %}`},
		{`{% with .foo %}{% end with %}`},
		{`{% with /. %}{% end with %}`},
		{`{% evoke foo %}`},
		{`{% evoke foo . %}`},
		{`{% evoke foo /. %}`},
		{`{% evoke foo .ctx %}`},
		{`{% evoke foo .c.t.x %}`},
		{`{% evoke foo /.ctx %}`},
		{`{% evoke foo /.c.t.x %}`},
		{`{% range . %}{% end range %}`},
		{`{% range . as _ _ %}{% end range %}`},
		{`{% range . as k v %}{% end range %}`},
		{`{% range . as _ v %}{% end range %}`},
		{`{% range . as k _ %}{% end range %}`},
		{`{% . %}`},
		{`{% /. %}`},
		{`{% /.foo %}`},
		{`{% $.foo %}`},
		{`{% call foo %}`},
		{`{% call foo /. %}`},
	})
}

func TestParserPeek(t *testing.T) {
	in := make(chan token)
	go func() {
		for i := 0; i < 10; i++ {
			in <- token{tokenType(i), nil, 0, 0}
		}
	}()
	p := &parser{
		in:   in,
		end:  tokenNoneType,
		errd: tokenNone,
	}
	ex := []int{0, 1, 1, 1, 2, 3, 3, 3, 3, 4, 5, 6}
	got := []int{}
	got = append(got, int(p.next().typ)) // 0 -> 1
	got = append(got, int(p.next().typ)) // 1 -> 2
	p.backup()                           // 2 -> 1 *
	got = append(got, int(p.peek().typ)) // 1 -> 1
	got = append(got, int(p.next().typ)) // 1 -> 2
	got = append(got, int(p.next().typ)) // 2 -> 3
	got = append(got, int(p.peek().typ)) // 3 -> 3
	got = append(got, int(p.peek().typ)) // 3 -> 3
	got = append(got, int(p.next().typ)) // 3 -> 4
	p.backup()                           // 4 -> 3 *
	got = append(got, int(p.next().typ)) // 3 -> 4
	got = append(got, int(p.next().typ)) // 4 -> 5
	got = append(got, int(p.next().typ)) // 5 -> 6
	got = append(got, int(p.next().typ)) // 6 -> 7

	if !reflect.DeepEqual(ex, got) {
		t.Fatalf("Expected %v got %v", ex, got)
	}
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
