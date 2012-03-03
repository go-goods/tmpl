package tmpl

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type executer interface {
	fmt.Stringer
	Execute(io.Writer, *context) error
}

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
	fmt.Fprintln(&buf, "[")
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

type executeBlockValue struct {
	ident string
	executer
}

type executeBlockDesc struct {
	ident string
	ctx   valueType
}

func (e *executeBlockDesc) Execute(w io.Writer, c *context) (err error) {
	//ask the context for the most up to date executer
	ex := c.GetBlock(e.ident)
	return ex.Execute(w, c)
}

func (e *executeBlockDesc) String() string {
	return fmt.Sprintf("[block %s %v]", e.ident, e.ctx)
}

type executeWith struct {
	ctx valueType
	ex  executer
}

func (e *executeWith) Execute(w io.Writer, c *context) (err error) {
	c.Push(e.ctx.Value(c))
	defer c.Pop()
	return e.ex.Execute(w, c)
}

func (e *executeWith) String() string {
	return fmt.Sprintf("[with %s] %s", e.ctx, e.ex)
}

type executeRange struct {
	iter valueType
	ex   executer
}

func (e *executeRange) Execute(w io.Writer, c *context) (err error) {
	//TODO: have to reflect on the value in order to range it
	return nil
}

func (e *executeRange) String() string {
	return fmt.Sprintf("[range %s] %s", e.iter, e.ex)
}

func truthy(val interface{}) bool {
	//returns if the value is "truthy" like nonzero, nonempty, etc.
	//TODO: reflect on the value and figure it out
	return true
}

type executeIf struct {
	cond valueType
	succ executer
	fail executer
}

func (e *executeIf) Execute(w io.Writer, c *context) (err error) {
	t := truthy(e.cond.Value(c))
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
		return fmt.Sprintf("[if else %s] %s | %s", e.cond, e.succ, e.fail)
	}
	return fmt.Sprintf("[if %s] %s", e.cond, e.succ)
}
