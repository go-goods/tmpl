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

type executeBlock struct {
	ident string
	ctx   valueType
	ex    executer
}

func (e *executeBlock) Execute(w io.Writer, c *context) (err error) {
	return nil
}

func (e *executeBlock) String() string {
	return fmt.Sprintf("[block %s %v] %s", e.ident, e.ctx, e.ex)
}

type executeWith struct {
	ctx valueType
	ex  executer
}

func (e *executeWith) Execute(w io.Writer, c *context) (err error) {
	return nil
}

func (e *executeWith) String() string {
	return fmt.Sprintf("[with %s] %s", e.ctx, e.ex)
}

type executeRange struct {
	iter valueType
	ex   executer
}

func (e *executeRange) Execute(w io.Writer, c *context) (err error) {
	return nil
}

func (e *executeRange) String() string {
	return fmt.Sprintf("[range %s] %s", e.iter, e.ex)
}

type executeIf struct {
	cond valueType
	succ executer
	fail executer
}

func (e *executeIf) Execute(w io.Writer, c *context) (err error) {
	return nil
}

func (e *executeIf) String() string {
	if e.fail != nil {
		return fmt.Sprintf("[if else %s] %s | %s", e.cond, e.succ, e.fail)
	}
	return fmt.Sprintf("[if %s] %s", e.cond, e.succ)
}
