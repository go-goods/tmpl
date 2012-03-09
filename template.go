package tmpl

import (
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sync"
)

//Mode is a type that represents one of two modes, Production or Development.
//See CompileMode for details.
type Mode bool

//String prints the mode in a human readable format.
func (m Mode) String() string {
	if bool(m) {
		return "Production"
	}
	return "Development"
}

const (
	Development Mode = true
	Production  Mode = false
)

var (
	modeChan   = make(chan Mode)
	modeChange = make(chan Mode)
	flocks     = newFileLock()
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

//CompileMode sets the compilation mode for the package. In Development mode,
//templates read in and compile each file it needs to execute every time it needs
//to execute, always getting the most recent changes. In Production mode, templates
//read and compile each file they need only the first time, caching the results
//for subsequent Execute calls. By default, the package is in Production mode.
func CompileMode(mode Mode) {
	modeChange <- mode
}

var cache = map[string]*parseTree{}

func newTemplate(file string) *Template {
	return &Template{
		base: file,
	}
}

type funcDecl struct {
	name string
	val  reflect.Value
}

//Template is the type that represents a template. It is created by using the
//Parse function and dependencies are attached through Blocks and Call.
type Template struct {
	//base and globs represent work to be done
	base      string
	globs     []string
	tempglobs []string
	funcs     []funcDecl

	compileLk sync.Mutex

	//our parse tree
	tree *parseTree
}

//Blocks attaches all of the block definitions in files that match the glob 
//patterns to the template for every Execute call so the base template can
//evoke them.
func (t *Template) Blocks(globs ...string) *Template {
	t.globs = append(t.globs, globs...)
	return t
}

//Call attaches a function to the template under the specified name for every
//Execute call so the base template can call them. The second argument must
//be a function, or Call will panic.
func (t *Template) Call(name string, fnc interface{}) *Template {
	rv := reflect.ValueOf(fnc)
	if rv.Kind() != reflect.Func {
		panic(fmt.Errorf("%q is not a function.", fnc))
	}
	t.funcs = append(t.funcs, funcDecl{name, rv})
	return t
}

func (t *Template) compile(mode Mode) (err error) {
	//figure out what work needs to be done
	if t.tree == nil || mode == Development {
		err = t.updateBase(mode)
		if err != nil {
			return
		}
	}
	if len(t.globs) > 0 {
		err = t.updateGlobs(t.globs, mode)
		if err != nil {
			return
		}
	}
	if len(t.funcs) > 0 {
		for _, decl := range t.funcs {
			t.tree.context.funcs[decl.name] = decl.val
		}
		t.funcs = nil
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
	flocks.Lock(abs)
	defer flocks.Unlock(abs)

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

//Execute runs the template with the specified context attaching all the block
//definitions in the files that match the given globs sending the output to
//w. Any errors during the compilation of any files that have to be compiled
//(see the discussion on Modes) or during the execution of the template are
//returned.
func (t *Template) Execute(w io.Writer, ctx interface{}, globs ...string) (err error) {
	t.compileLk.Lock()

	mode := <-modeChan
	if mode == Development || t.tree == nil {
		if t.tree != nil {
			t.tree.context.clear()
		}
		t.compile(mode)
	}

	t.tree.context.dup()
	defer t.tree.context.restore()
	if len(globs) > 0 {
		if err = t.updateGlobs(globs, mode); err != nil {
			return
		}
	}
	t.compileLk.Unlock()

	//execute!
	return t.tree.Execute(w, ctx)
}

//Parse creates a new Template with the specified file acting as the base
//template.
func Parse(file string) (t *Template) {
	t = newTemplate(file)
	return
}
