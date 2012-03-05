package tmpl

import (
	"reflect"
	"testing"
)

var code = []byte(`
	literal
	{% call func .foo$bar .bar$$baz..foo %}
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
	{% range . as _ val %}{% with .val %}{% . %}{% end with %}{% end range %}`)

func BenchmarkParseSpeed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parse(lex(code))
	}
}

func TestParseNestedBlocks(t *testing.T) {
	_, err := parse(lex([]byte(`{% block foo %} foo {% block bar %} bar {% end block %} foo {% end block %}`)))
	if err == nil {
		t.Errorf("Expected error parsing nested blocks.")
	}
	_, err = parse(lex([]byte(`{%block foo%}{%with .%}{%block bar%}{%end block%}{%end with%}{%end block%}`)))
	if err == nil {
		t.Errorf("Expected error parsing nested blocks.")
	}

}

func TestParseRedefineBlock(t *testing.T) {
	_, err := parse(lex([]byte(`{% block foo %}{% end block %}{% block foo %}{% end block %}`)))
	if err == nil {
		t.Errorf("Expected error redefining a block")
	}
}

func TestParseBasic(t *testing.T) {
	ch := lex(code)
	out := make(chan token)
	go func() {
		for i := range ch {
			// t.Log(i)
			out <- i
		}
		close(out)
	}()
	tree, err := parse(out)
	if err != nil {
		t.Fatal(err)
	}
	_ = tree
	// t.Log(tree)
}

func TestParsePeek(t *testing.T) {
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
