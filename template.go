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
	if err != nil {
		return
	}
	tree.context.setFile(file)
	return
}

func newTemplate(file string) *Template {
	return &Template{
		base: file,
	}
}

type Template struct {
	//cache of compiled files
	cache map[string]*parseTree

	//base and globs represent work to be done
	base      string
	globs     []string
	tempglobs []string

	//if we have compiled
	compiled bool

	//our parse tree
	tree *parseTree
}

func (t *Template) Blocks(globs ...string) *Template {
	t.globs = append(t.globs, globs...)
	return t
}

func (t *Template) Compile() (err error) {
	err = t.compile()
	if err != nil {
		return
	}
	t.compiled = true
	return err
}

func (t *Template) compile() (err error) {
	//figure out what work needs to be done
	if t.base != "" {
		err = t.updateBase()
		if err != nil {
			return
		}
		t.base = ""
	}
	if len(t.globs) > 0 {
		err = t.updateGlobs(t.globs)
		if err != nil {
			return
		}
		t.globs = nil
	}
	if len(t.tempglobs) > 0 {
		err = t.updateGlobs(t.tempglobs)
		if err != nil {
			return
		}
		t.tempglobs = nil
	}
}

//treeFor grabs the parseTree for the specified absolute path, grabbing it from
//the cache if t.compiled is true
func (t *Template) treeFor(abs string) (tree *parseTree, err error) {
	if t.compiled {
		//check for the cache
		if tr, ex := t.cache[abs]; ex {
			tree = tr
			return
		}
	}
	tree, err = parseFile(abs)
	if err != nil {
		return
	}
	t.cache[abs] = tree
}

func (t *Template) updateBase() (err error) {
	abs, err := filepath.Abs(t.base)
	if err != nil {
		return
	}
	t.tree, err = t.treeFor(abs)
	return
}

func (t *Template) updateGlobs(globs []string) (err error) {
	for _, glob := range globs {
		err = t.updateGlob(glob)
		if err != nil {
			return
		}
	}
}

func (t *Template) updateGlob(glob string) (err error) {
	files, err := filepath.Glob(glob)
	if err != nil {
		return
	}
	for _, file := range files {
		err = t.loadBlocks(file)
		if err != nil {
			return
		}
	}
}

func (t *Template) loadBlocks(file string) (err error) {
	abs, err := filepath.Abs(file)
	if err != nil {
		return
	}
	tree, err := t.treeFor(abs)
	if err != nil {
		return
	}
	err = t.updateBlocks(file, tree.context.blocks)
	return
}

func (t *Template) updateBlocks(file string, blocks []*executeBlockValue) (err error) {
	tblk := t.tree.context.blocks
	for id, blk := range blocks {
		if bl, ex := tblk[id]; ex {
			err = fmt.Errorf("%q: %q already exists from %q", file, id, blk.file)
			return
		}
		blk.file = file
		tblk[id] = blk
	}
	return
}

func (t *Template) Execute(w io.Writer, ctx interface{}, globs ...string) (err error) {
	//backup the context
	t.tree.context.dup()
	defer t.tree.context.restore()

	//add in our temporary blocks
	t.tempglobs = append(t.tempglobs, globs...)

	err = t.compile()
	if err != nil {
		return
	}

	//execute!
	return t.tree.Execute(w, ctx)
}

func Parse(file string) (t *Template) {
	t = newTemplate(file)
	return
}
