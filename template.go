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
		base:  file,
		dirty: true,
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
	base  string
	globs []string
	funcs []funcDecl
	dirty bool

	compileLk sync.RWMutex

	//our parse tree
	tree *parseTree
}

//Blocks attaches all of the block definitions in files that match the glob 
//patterns to the template for every Execute call so the base template can
//evoke them.
func (t *Template) Blocks(globs ...string) *Template {
	t.globs = append(t.globs, globs...)
	t.dirty = true
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
	t.dirty = true
	return t
}

func (t *Template) compile(mode Mode) (err error) {
	if err = t.updateBase(mode); err != nil {
		return
	}
	if err = t.updateGlobs(t.globs, mode); err != nil {
		return
	}
	for _, decl := range t.funcs {
		t.tree.context.funcs[decl.name] = decl.val
	}
	t.dirty = false
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
	mode := <-modeChan
	if mode == Development || t.dirty {
		//grab the compile lock
		t.compileLk.Lock()

		//unset the tree and compile it
		t.tree = nil
		if err = t.compile(mode); err != nil {
			t.compileLk.Unlock()
			return
		}
		t.tree.context.dup()

		//if we have temp things
		if len(globs) > 0 {
			//set up a restore
			defer t.tree.context.restore()
			//load them in
			if err = t.updateGlobs(globs, mode); err != nil {
				t.compileLk.Unlock()
				return
			}
		}
		//done compiling!
		t.compileLk.Unlock()
	} else {
		//we arent dirty or in dev mode, but we could have temp globs
		if len(globs) > 0 {
			//so grab the compile lock
			t.compileLk.Lock()
			//set up a restore
			defer t.tree.context.restore()
			//load them in
			if err = t.updateGlobs(globs, mode); err != nil {
				t.compileLk.Unlock()
				return
			}
			//done compiling!
			t.compileLk.Unlock()
		}
	}

	//execute!
	t.compileLk.RLock()
	defer t.compileLk.RUnlock()
	return t.tree.Execute(w, ctx)
}

//Parse creates a new Template with the specified file acting as the base
//template.
func Parse(file string) (t *Template) {
	t = newTemplate(file)
	return
}
