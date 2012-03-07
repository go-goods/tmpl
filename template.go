package tmpl

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
)

func parseFile(file string) (tree *parseTree, err error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return
	}
	tree, err = parse(lex(data))
	tree.context.setFile(file)
	return
}

func newTemplate(tree *parseTree) *Template {
	return &Template{
		tree: tree,
	}
}

type Template struct {
	tree *parseTree
}

func (t *Template) attachGlob(glob string) (err error) {
	m, err := filepath.Glob(glob)
	if err != nil {
		return
	}

	for _, file := range m {
		err = t.attachBlocks(file)
		if err != nil {
			return
		}
	}
	return
}

func (t *Template) attachBlocks(file string) (err error) {
	tree, err := parseFile(file)
	if err != nil {
		return
	}
	//grab the blocks out
	for key, val := range tree.context.blocks {
		if bl, ex := t.tree.context.blocks[key]; ex {
			return fmt.Errorf("%q: Block named %q already loaded from %q", file, bl.ident, bl.file)
		}
		t.tree.context.blocks[key] = val
	}
	return
}

func (t *Template) Blocks(globs ...string) (err error) {
	for _, glob := range globs {
		err = t.attachGlob(glob)

		if err != nil {
			return
		}
	}
	return
}

func (t *Template) Execute(w io.Writer, ctx interface{}, globs ...string) (err error) {
	//backup the context
	t.tree.context.dup()
	defer t.tree.context.restore()

	//add in our temporary blocks
	err = t.Blocks(globs...)
	if err != nil {
		return
	}

	//execute!
	return t.tree.Execute(w, ctx)
}

func Parse(file string) (t *Template, err error) {
	tree, err := parseFile(file)
	if err != nil {
		return
	}
	t = newTemplate(tree)
	return
}

func MustParse(file string) (t *Template) {
	t, err := Parse(file)
	if err != nil {
		panic(err)
	}
	return
}
