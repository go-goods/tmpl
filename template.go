package tmpl

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
)

type Mode bool

func (m Mode) String() string {
	if bool(m) {
		return "Production"
	}
	return "Development"
}

const (
	Development Mode = false
	Production  Mode = true
)

var (
	modeChan   = make(chan Mode)
	modeChange = make(chan Mode)
)

func init() {
	go modeSpitter()
}

func modeSpitter() {
	mode := Development
	for {
		select {
		case modeChan <- mode:
		case mode = <-modeChange:
		}
	}
}

func CompileMode(mode Mode) {
	modeChange <- mode
}

var cache = map[string]*parseTree{}

func newTemplate(file string) *Template {
	return &Template{
		base: file,
	}
}

type Template struct {
	//base and globs represent work to be done
	base      string
	globs     []string
	tempglobs []string

	//our parse tree
	tree *parseTree
}

func (t *Template) Blocks(globs ...string) *Template {
	t.globs = append(t.globs, globs...)
	return t
}

func (t *Template) compile() (err error) {
	mode := <-modeChan
	//figure out what work needs to be done
	if t.base != "" {
		err = t.updateBase(mode)
		if err != nil {
			return
		}
		if mode == Production {
			t.base = ""
		}
	}
	if len(t.globs) > 0 {
		err = t.updateGlobs(t.globs, mode)
		if err != nil {
			return
		}
		if mode == Production {
			t.globs = nil
		}
	}
	if len(t.tempglobs) > 0 {
		err = t.updateGlobs(t.tempglobs, mode)
		if err != nil {
			return
		}
		t.tempglobs = nil
	}
	return
}

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

//treeFor grabs the parseTree for the specified absolute path, grabbing it from
//the cache if t.compiled is true
func (t *Template) treeFor(abs string, mode Mode) (tree *parseTree, err error) {
	if mode == Production {
		//check for the cache
		if tr, ex := cache[abs]; ex {
			tree = tr
			return
		}
	}
	tree, err = parseFile(abs)
	if err != nil {
		return
	}
	cache[abs] = tree
	return
}

func (t *Template) updateBase(mode Mode) (err error) {
	abs, err := filepath.Abs(t.base)
	if err != nil {
		return
	}
	t.tree, err = t.treeFor(abs, mode)
	return
}

func (t *Template) updateGlobs(globs []string, mode Mode) (err error) {
	for _, glob := range globs {
		err = t.updateGlob(glob, mode)
		if err != nil {
			return
		}
	}
	return
}

func (t *Template) updateGlob(glob string, mode Mode) (err error) {
	files, err := filepath.Glob(glob)
	if err != nil {
		return
	}
	for _, file := range files {
		err = t.loadBlocks(file, mode)
		if err != nil {
			return
		}
	}
	return
}

func (t *Template) loadBlocks(file string, mode Mode) (err error) {
	abs, err := filepath.Abs(file)
	if err != nil {
		return
	}
	tree, err := t.treeFor(abs, mode)
	if err != nil {
		return
	}
	err = t.updateBlocks(file, tree.context.blocks)
	return
}

func (t *Template) updateBlocks(file string, blocks map[string]*executeBlockValue) (err error) {
	tblk := t.tree.context.blocks
	for id, bl := range blocks {
		if _, ex := tblk[id]; ex {
			err = fmt.Errorf("%q: %q already exists from %q", file, id, bl.file)
			return
		}
		bl.file = file
		tblk[id] = bl
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
