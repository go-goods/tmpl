package tmpl

import (
	"bytes"
	"fmt"
	"io"
	"reflect"
	"regexp"
	"strings"
)

type executer interface {
	fmt.Stringer
	Execute(io.Writer, *context) error
}

// ****************
// * Execute List *
// ****************

var whitespace = regexp.MustCompile(`^\s+$`)

type executeList []executer

func (e executeList) Execute(w io.Writer, c *context) (err error) {
	for _, ex := range e {
		if ex == nil {
			return fmt.Errorf("unexpected nil in execute list")
		}
		err = ex.Execute(w, c)
		if err != nil {
			return
		}
	}
	return
}

func (e executeList) String() string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "[list")
	for _, ex := range e {
		if ex != nil {
			fmt.Fprintf(&buf, "\t%s\n", strings.Replace(ex.String(), "\n", "\n\t", -1))
		} else {
			fmt.Fprint(&buf, "\tnil\n")
		}
	}
	fmt.Fprint(&buf, "]")
	return buf.String()
}

func (e *executeList) Push(ex executer) {
	*e = append(*e, ex)
}

func (e *executeList) compact() {
	//take if statements that are always true and replace them
	e.substituteTrueIf()
	//take runs of constant expressions and simply them
	e.combineConstant()
	//drops any constants that are entirely whitespace
	e.dropWhitespace()
}

func (e *executeList) substituteTrueIf() {
	for idx, ex := range *e {
		if eIf, ok := ex.(*executeIf); ok {
			//if it is a constant if that can be known at compile time
			if val, isConst := eIf.constValue(); isConst {
				(*e)[idx] = val
			}
		}
	}
	//make a secondary list to copy into without nils
	cl := make(executeList, 0, len(*e))
	for _, ex := range *e {
		if ex != nil {
			cl = append(cl, ex)
		}
	}
	*e = cl
	return
}
func (e *executeList) combineConstant() {
	//make a secondary list to copy in folded constants
	cl := make(executeList, 0, len(*e))
	//run through looking for runs of constant values
	for i := 0; i < len(*e); i++ {
		//grab the  current element
		ex := (*e)[i]

		//if its a constantValue attempt to fold it
		if co, ok := ex.(constantValue); ok {
			i++ //look at the next element

			//while we dont run off the end of the array
			for ; i < len(*e); i++ {
				//check if we have a constant value
				ne, ok := (*e)[i].(constantValue)
				if !ok {
					//backup if we dont
					i--
					break
				}
				//append the constant to the previous one
				co.Append([]byte(ne))
			}
			//set our executer to the folded constant
			ex = co
		}
		//append our element
		cl = append(cl, ex)
	}
	*e = cl
	return
}

func (e *executeList) dropWhitespace() {
	//make a secondary list to copy in non whitespace
	cl := make(executeList, 0, len(*e))
	for _, ex := range *e {
		//if its a whitespace constant value, skip it
		if co, ok := ex.(constantValue); ok && whitespace.Match([]byte(co)) {
			continue
		}
		//otherwise add it to our little list
		cl = append(cl, ex)
	}
	//set the value to our new list
	*e = cl
	return
}

// ***********************
// * Execute Block Value *
// ***********************

type executeBlockValue struct {
	ident string
	ex    executer
}

func (e *executeBlockValue) Execute(w io.Writer, c *context) (err error) {
	if e.ex == nil {
		return
	}
	return e.ex.Execute(w, c)
}

func (e *executeBlockValue) String() string {
	return e.ex.String()
}

// *****************
// * Execute Evoke *
// *****************

type executeEvoke struct {
	ident string
	ctx   *selectorValue
}

func (e *executeEvoke) Execute(w io.Writer, c *context) (err error) {
	//ask the context for the most up to date executer
	ex := c.getBlock(e.ident)
	if ex == nil {
		return fmt.Errorf("No block by the name %s", e.ident)
	}

	//set up our context
	if e.ctx != nil {
		defer c.setStack(c.stack.dup())
		if err = c.cd(e.ctx); err != nil {
			return
		}
	}

	return ex.Execute(w, c)
}

func (e *executeEvoke) String() string {
	return fmt.Sprintf("[block %s %v]", e.ident, e.ctx)
}

// ****************
// * Execute With *
// ****************

type executeWith struct {
	ctx *selectorValue
	ex  executer
}

func (e *executeWith) Execute(w io.Writer, c *context) (err error) {
	//check that our executer exists
	if e.ex == nil {
		return
	}

	//set up our context
	defer c.setStack(c.stack.dup())
	if err = c.cd(e.ctx); err != nil {
		return
	}

	return e.ex.Execute(w, c)
}

func (e *executeWith) String() string {
	return fmt.Sprintf("[with %s] %s", e.ctx, e.ex)
}

// *****************
// * Execute Range *
// *****************

type executeRange struct {
	iter     *selectorValue
	ex       executer
	key, val token
}

func (e *executeRange) Execute(w io.Writer, c *context) (err error) {
	//TODO: have to reflect on the value in order to range it
	//be sure to look at e.key and e.val to set/unset the vars
	it, err := e.iter.Value(c)
	if err != nil {
		return
	}

	var kstr, vstr string
	if s := string(e.key.dat); s != "" && s != "_" {
		kstr = c.stack.StringWith([]string{s})
	}
	if s := string(e.val.dat); s != "" && s != "_" {
		vstr = c.stack.StringWith([]string{s})
	}

	//check to see if we can iterate over it
	rv := reflect.ValueOf(it)
	switch rv.Kind() {
	case reflect.Map:
		err = e.rangeMap(w, c, rv, kstr, vstr)
	case reflect.Slice, reflect.Array:
		err = e.rangeSlice(w, c, rv, kstr, vstr)
	case reflect.Struct:
		err = e.rangeStruct(w, c, rv, kstr, vstr)
	default:
		err = fmt.Errorf("%s is a %v, a non iterable type", e.iter, rv.Kind())
	}

	c.unsetAt(kstr)
	c.unsetAt(vstr)

	return
}

func (e *executeRange) String() string {
	return fmt.Sprintf("[range %s] %s", e.iter, e.ex)
}

func (e *executeRange) rangeMap(w io.Writer, c *context, v reflect.Value, vstr, kstr string) (err error) {
	for _, key := range v.MapKeys() {
		c.setAt(kstr, indirect(key).Interface())
		c.setAt(vstr, indirect(v.MapIndex(key)).Interface())
		err = e.ex.Execute(w, c)
	}
	return
}

func (e *executeRange) rangeSlice(w io.Writer, c *context, v reflect.Value, vstr, kstr string) (err error) {
	for i := 0; i < v.Len(); i++ {
		c.setAt(kstr, i)
		c.setAt(vstr, indirect(v.Index(i)).Interface())
		err = e.ex.Execute(w, c)
	}
	return
}

func (e *executeRange) rangeStruct(w io.Writer, c *context, v reflect.Value, vstr, kstr string) (err error) {
	for i := 0; i < v.NumField(); i++ {
		c.setAt(kstr, i)
		c.setAt(vstr, indirect(v.Field(i)).Interface())
		err = e.ex.Execute(w, c)
	}
	return
}

// **************
// * Execute If *
// **************

type executeIf struct {
	cond valueType
	succ executer
	fail executer
}

func (e *executeIf) constValue() (ex executer, isConst bool) {
	if isConstantValue(e.cond) {
		isConst = true

		//grab the value. there should never be errors getting the value of a
		//constant
		v, err := e.cond.Value(nil)
		if err != nil {
			panic(err)
		}

		if truthy(v) {
			ex = e.succ
		} else {
			ex = e.fail
		}
	}
	return
}

func (e *executeIf) Execute(w io.Writer, c *context) (err error) {
	v, err := e.cond.Value(c)
	if err != nil {
		return
	}
	t := truthy(v)
	if t {
		return e.succ.Execute(w, c)
	}
	if e.fail != nil {
		return e.fail.Execute(w, c)
	}
	return nil
}

func (e *executeIf) String() string {
	if e.fail != nil {
		return fmt.Sprintf("[if else %s] %v | %v", e.cond, e.succ, e.fail)
	}
	return fmt.Sprintf("[if %s] %v", e.cond, e.succ)
}

// truthy returns whether the value is 'true', in the sense of not the zero of its type,
// and whether the value has a meaningful truth value.
func truthy(i interface{}) (truth bool) {
	val := reflect.ValueOf(i)
	if !val.IsValid() {
		// Something like var x interface{}, never set. It's a form of nil.
		return false
	}
	switch val.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		truth = val.Len() > 0
	case reflect.Bool:
		truth = val.Bool()
	case reflect.Complex64, reflect.Complex128:
		truth = val.Complex() != 0
	case reflect.Chan, reflect.Func, reflect.Ptr, reflect.Interface:
		truth = !val.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		truth = val.Int() != 0
	case reflect.Float32, reflect.Float64:
		truth = val.Float() != 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		truth = val.Uint() != 0
	case reflect.Struct:
		truth = true // Struct values are always true.
	}
	return
}
