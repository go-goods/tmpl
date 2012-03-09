package tmpl_test

import (
	"github.com/goods/tmpl"
	"io/ioutil"
)

var w ioutil.Discard

func ExampleCompileMode() {
	//turn on Production mode
	tmpl.CompileMode(tmpl.Production)
	//turn on Development mode
	tmpl.CompileMode(tmpl.Development)
}

func ExampleParse() {
	t := tmpl.Parse("tmpls/base.tmpl")
	if err := t.Execute(w, nil); err != nil {
		panic(err)
	}
}

func ExampleTemplate_Blocks() {
	t := tmpl.Parse("tmpls/base.tmpl")

	//attach the blocks we need for every Execute call
	t.Blocks("tmpls/content.block", "tmpls/base/*.block")
	t.Blocks("tmpls/another.block")

	//now the block definitions in the specified files will be used for evoke
	//calls in the base template.
	if err := t.Execute(w, nil); err != nil {
		panic(err)
	}
}

func ExampleTemplate_Call() {
	t := tmpl.Parse("tmpls/base.tmpl")

	//attach the functions that will be available for every Execute call
	t.Call("foo", func() string {
		return "A foo!"
	})
	t.Call("bar", func(a int, x string) (string, string) {
		return "a foo", "and a bar!"
	})

	//now the functions "foo" and "bar" will be available in for call in the
	//base template.
	if err := t.Execute(w, nil); err != nil {
		panic(err)
	}
}

func ExampleTemplate_Execute() {
	t := tmpl.Parse("tmpls/base.tmpl")

	//create a context for the call
	type d map[string]interface{}
	ctx := d{
		"foo": "bar",
		"baz": []string{"and a one", "and a two", "and a"},
		"things": d{
			"one":   1,
			"two":   2,
			"three": 3,
		},
	}

	//call the base template on that context, and load in some blocks only for
	//this Execute call.
	if err := t.Execute(w, ctx, "tmpls/site_page/*.block"); err != nil {
		panic(err)
	}
}
