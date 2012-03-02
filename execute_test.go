package tmpl

import (
	"bytes"
	"testing"
)

func TestExecuteLiteral(t *testing.T) {
	const lit = `this is just a literal`
	tree, err := parse(lex([]byte(lit)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if buf.String() != lit {
		t.Fatalf("\nGot %q\nExp %q", buf.String(), lit)
	}
}

func TestExecutePositiveIf(t *testing.T) {
	const lit = `{% if 1 %}test{% end if %}`
	tree, err := parse(lex([]byte(lit)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, nil); err != nil {
		t.Fatal(err)
	}
	if got, ex := buf.String(), `test`; got != ex {
		t.Fatalf("\nGot %q\nExp %q", got, ex)
	}
}

func TestDefaultBlock(t *testing.T) {
	const lit = `{% block foo %}test{% end block %}`
	tree, err := parse(lex([]byte(lit)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, &context{}); err != nil {
		t.Fatal(err)
	}
	if got, ex := buf.String(), `test`; got != ex {
		t.Fatalf("\nGot %q\nExp %q", got, ex)
	}
}

func TestLiteralChain(t *testing.T) {
	const lit = `t{%%}e{%%}s{%%}t{%%}`
	tree, err := parse(lex([]byte(lit)))
	if err != nil {
		t.Fatal(err)
	}
	var buf bytes.Buffer
	if err := tree.Execute(&buf, &context{}); err != nil {
		t.Fatal(err)
	}
	if got, ex := buf.String(), `test`; got != ex {
		t.Fatalf("\nGot %q\nExp %q", got, ex)
	}
}
