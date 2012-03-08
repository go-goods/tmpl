package tmpl

import (
	"bytes"
	"io/ioutil"
	"testing"
)

func TestExecuteListString(t *testing.T) {
	l := executeList{
		nil,
		executeList{nil, nil, nil},
		nil,
	}
	l.Push(nil)
	if l.String() != "[list\n\tnil\n\t[list\n\t\tnil\n\t\tnil\n\t\tnil\n\t]\n\tnil\n\tnil\n]" {
		t.Error("didn't nest right")
	}
}

func TestExecuteNoContext(t *testing.T) {
	cases := []struct {
		templ  string
		expect string
	}{
		{`this is just a literal`, `this is just a literal`},
		{`{% block foo %}test{% end block %}{% evoke foo %}`, `test`},
		{`{# foo #}test`, `test`},
		{`{# foo #}test{# bar baz #}`, `test`},
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

func TestExecuteSimpleBlock(t *testing.T) {
	tree, err := parse(lex([]byte(`{% block foo %}{% . %}{% end block %}{% evoke foo .foo %}`)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, d{"foo": "foo"}); err != nil {
		t.Fatal(err)
	}
	if g := buf.String(); g != "foo" {
		t.Fatalf("\nGot %q\nExp %q", g, "foo")
	}
}

func TestExecuteSimpleWith(t *testing.T) {
	tree, err := parse(lex([]byte(`{% with .foo %}{% . %}{% end with %}{% .foo %}`)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, d{"foo": "foo"}); err != nil {
		t.Fatal(err)
	}
	if g := buf.String(); g != "foofoo" {
		t.Fatalf("\nGot %q\nExp %q", g, "foofoo")
	}
}

func TestExecuteEvokeNoBlock(t *testing.T) {
	tree, err := parse(lex([]byte(`{% evoke foo %}`)))
	if err != nil {
		t.Fatal(err)
	}

	if err := tree.Execute(ioutil.Discard, nil); err == nil {
		t.Fatal("Expected error evoking a block with no definition")
	}
}

func TestExecuteIfConstanVal(t *testing.T) {
	var sentinal executer = intValue(2)
	cases := []*executeIf{
		{intValue(1), sentinal, nil},
		{floatValue(1), sentinal, nil},
		{constantValue(`foo`), sentinal, nil},
		{intValue(0), nil, sentinal},
		{floatValue(0), nil, sentinal},
		{constantValue(``), nil, sentinal},
	}
	for _, i := range cases {
		if e, isConst := i.constValue(); !isConst || e != sentinal {
			t.Fatal("Expected const sentinal on", i)
		}
	}
}

func TestExecuteListSubstituteIf(t *testing.T) {
	var sentinal executer = intValue(2)
	e := executeList{
		&executeIf{intValue(1), sentinal, nil},
		&executeIf{floatValue(1), sentinal, nil},
		&executeIf{constantValue(`foo`), sentinal, nil},
		&executeIf{intValue(0), nil, sentinal},
		&executeIf{floatValue(0), nil, sentinal},
		&executeIf{constantValue(``), nil, sentinal},
	}
	b := len(e)
	e.substituteTrueIf()
	if b != len(e) {
		t.Fatalf("Lost some items %d -> %d", b, len(e))
	}
	for idx, ex := range e {
		if i, ok := ex.(intValue); !ok || i != sentinal {
			t.Errorf("item %d fails", idx)
		}
	}
}

func TestExecuteListCombineConstant(t *testing.T) {
	e := executeList{
		constantValue(`foo`),
		constantValue(`bar`),
		constantValue(`baz`),
		intValue(2),
		constantValue(`foo`),
		constantValue(`bar`),
		constantValue(`baz`),
	}
	e.combineConstant()
	if len(e) != 3 {
		t.Fatal("Expected 3. Got %d\n%v", len(e), e)
	}
	if g := string(e[0].(constantValue)); g != `foobarbaz` {
		t.Errorf("Value incorrect. %q", g)
	}
	if g := int64(e[1].(intValue)); g != 2 {
		t.Errorf("Value incorrect. %q", g)
	}
	if g := string(e[2].(constantValue)); g != `foobarbaz` {
		t.Errorf("Value incorrect. %q", g)
	}
}

func TestExecuteListDropWhitespace(t *testing.T) {
	e := executeList{
		constantValue(`foo`),
		constantValue(` `),
		constantValue(`foo`),
		constantValue(`	`),
		constantValue(`foo`),
		constantValue("\r\n"),
		constantValue(`foo`),
	}
	e.dropWhitespace()
	if len(e) != 4 {
		t.Fatalf("Expected 4 got %d", len(e))
	}
	for i, cv := range e {
		if o, ok := cv.(constantValue); !ok || string(o) != "foo" {
			t.Errorf("%dth foo is %v", i, cv)
		}
	}
}
