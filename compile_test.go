package tmpl

import (
	"reflect"
	"testing"
)

var code = []byte(`
	literal
	{% call func .foo$bar .bar$$baz..foo 2.3 5e7 "boof" %}
	{% block baz .buff %}
		doof bood
		{% range call foob .bar %}
			ding dong
		{% end range %}
		{% with call foob .bar %}
			{% block butt %}
				dar fangle {% if .foo %}doo{% else %}no doo{% end if %}
			{% end block %}
		{% end with %}
	{% end block %}
	{% if "foo" %}
		always!
	{% end if %}{% if "foo" %}{% else %}doof{% end if %}
	{% with 25 %}{% . %}{% end with %}`)

func BenchmarkParseSpeed(b *testing.B) {
	for i := 0; i < b.N; i++ {
		parse(lex(code))
	}
}

func TestExecuteListString(t *testing.T) {
	l := executeList{
		nil,
		executeList{nil, nil, nil},
		nil,
	}
	l.Push(nil)
	if l.String() != "[\n\tnil\n\t[\n\t\tnil\n\t\tnil\n\t\tnil\n\t]\n\tnil\n\tnil\n]" {
		t.Error("didn't nest right")
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
