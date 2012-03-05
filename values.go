package tmpl

import (
	"bytes"
	"fmt"
	"io"
	"strconv"
)

type valueType interface {
	executer
	Value(*context) interface{}
}

// ***********
// * Helpers *
// ***********

func isConstantValue(v valueType) bool {
	switch v.(type) {
	case intValue, floatValue, constantValue:
		return true
	}
	return false
}

// *******************
// * Parsing Helpers *
// *******************

func isValueType(tok token) bool {
	switch tok.typ {
	case tokenStartSel, tokenCall, tokenValue, tokenNumeric:
		return true
	}
	return false
}

func numericToValue(tok token) (v valueType, err error) {
	if tok.typ != tokenNumeric {
		return nil, fmt.Errorf("expected numeric got %q", tok)
	}
	sval := string(tok.dat)
	i, err := strconv.ParseInt(sval, 10, 64)
	if err == nil {
		return intValue(i), nil
	}
	f, err := strconv.ParseFloat(sval, 64)
	if err == nil {
		return floatValue(f), nil
	}
	return
}

func consumeValue(p *parser) (valueType, error) {
	switch tok := p.next(); tok.typ {
	case tokenStartSel, tokenValue, tokenNumeric:
		p.backup()
		return consumeBasicValue(p)
	case tokenCall:
		return consumeCallValue(p)
	default:
		return nil, fmt.Errorf("Expected a value type got a %q", tok)
	}
	return nil, nil
}

func consumeCallValue(p *parser) (valueType, error) {
	//grab the name identifier
	name := p.next()
	if name.typ != tokenIdent {
		return nil, fmt.Errorf("Expected a %q got a %q", tokenIdent, name)
	}
	//grab values until p.peek() is a tokenClose
	values := []valueType{}
	for p.peek().typ != tokenClose {
		//consume a basic value
		val, err := consumeBasicValue(p)
		if err != nil {
			return nil, err
		}
		//append it
		values = append(values, val)
	}
	return callValue{name.dat, values}, nil
}

func consumeBasicValue(p *parser) (valueType, error) {
	switch tok := p.next(); tok.typ {
	case tokenStartSel:
		toks := p.acceptUntil(tokenEndSel)
		//consume the end sel
		if p.next().typ != tokenEndSel {
			return nil, fmt.Errorf("Expected a %q got a %q", tokenEndSel, p.curr)
		}
		return createSelector(toks), nil
	case tokenValue:
		return constantValue(tok.dat), nil
	case tokenNumeric:
		return numericToValue(tok)
	default:
		return nil, fmt.Errorf("Expected a value type got got a %q", tok)
	}
	return nil, nil
}

// ******************
// * Selector Value *
// ******************

type selectorValue []token

func (s selectorValue) Value(c *context) interface{} {
	return nil
}

func (s selectorValue) Execute(w io.Writer, c *context) (err error) {
	_, err = fmt.Fprint(w, s.Value(c))
	return
}

func (s selectorValue) String() string {
	var buf bytes.Buffer
	fmt.Fprint(&buf, "[selector")
	for _, tok := range s {
		fmt.Fprintf(&buf, " %s", tok)
	}
	fmt.Fprint(&buf, "]")
	return buf.String()
}

func createSelector(toks []token) selectorValue {
	return selectorValue(toks)
}

// **************
// * Call Value *
// **************

type callValue struct {
	name []byte
	args []valueType
}

func (s callValue) Value(c *context) interface{} {
	return nil
}

func (s callValue) Execute(w io.Writer, c *context) (err error) {
	_, err = fmt.Fprint(w, s.Value(c))
	return
}

func (s callValue) String() string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "[call %s", string(s.name))
	for _, v := range s.args {
		fmt.Fprintf(&buf, " %s", v)
	}
	fmt.Fprint(&buf, "]")
	return buf.String()
}

// ************************
// * CONSTANT VALUE TYPES *
// ************************

// *************
// * Int Value *
// *************

type intValue int64

func (s intValue) Value(c *context) interface{} {
	return int64(s)
}

func (s intValue) Execute(w io.Writer, c *context) (err error) {
	_, err = fmt.Fprint(w, s.Value(c))
	return
}

func (s intValue) String() string {
	return fmt.Sprintf("[int %v]", s.Value(nil))
}

func (s intValue) Byte() []byte {
	return []byte(s.String())
}

// ***************
// * Float Value *
// ***************

type floatValue float64

func (s floatValue) Value(c *context) interface{} {
	return float64(s)
}

func (s floatValue) Execute(w io.Writer, c *context) (err error) {
	_, err = fmt.Fprint(w, s.Value(c))
	return
}

func (s floatValue) String() string {
	return fmt.Sprintf("[float %f]", s.Value(nil))
}

func (s floatValue) Byte() []byte {
	return []byte(s.String())
}

// *************************
// * String Constant Value *
// *************************

type constantValue []byte

func (s constantValue) Value(c *context) interface{} {
	return string(s)
}

func (s constantValue) Execute(w io.Writer, c *context) (err error) {
	_, err = fmt.Fprint(w, s.Value(c))
	return
}

func (s constantValue) String() string {
	return fmt.Sprintf("[constant %q]", s.Value(nil))
}

func (s *constantValue) Append(p []byte) {
	*s = append(*s, p...)
}
