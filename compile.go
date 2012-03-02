package tmpl

import "fmt"

func Parse(data string) (string, error) {
	ch := lex([]byte(data))
	tree, err := parse(ch)
	if err != nil {
		return "", err
	}
	return tree.String(), nil
}

type parser struct {
	in     chan token
	out    chan executer
	err    chan error
	end    tokenType
	curr   token
	backed bool
	errd   token
}

type parseState func(*parser) parseState

func parse(toks chan token) (ex executer, err error) {
	return subParse(&parser{in: toks, errd: tokenNone}, tokenNoneType)
}

func (p *parser) run() {
	for state := parseText; state != nil; {
		state = state(p)
	}
	close(p.out)
}

func (p *parser) errorf(format string, args ...interface{}) parseState {
	p.err <- fmt.Errorf(format, args...)
	return nil
}

func (p *parser) errExpect(ex tokenType, got token) parseState {
	return p.errorf("Expected a %q got a %q", ex, got)
}

func (p *parser) unexpected(t token) parseState {
	// var stack [4096]byte
	// log.Println(t)
	// runtime.Stack(stack[:], false)
	// log.Println(string(stack[:]))
	return p.errorf("Unexpected %q", t)
}

func (p *parser) accept(tok tokenType) bool {
	if p.next().typ == tok {
		return true
	}
	p.backup()
	return false
}

func (p *parser) next() token {
	if p.backed {
		p.backed = false
		return p.curr
	}
	if p.errd.typ != tokenNoneType {
		return p.errd
	}
	p.curr = <-p.in
	switch p.curr.typ {
	case tokenEOF, tokenError:
		p.errd = p.curr
	}
	return p.curr
}

func (p *parser) backup() {
	if p.backed {
		panic("double backup")
	}
	p.backed = true
}

func (p *parser) peek() (t token) {
	t = p.next()
	p.backup()
	return
}

func (p *parser) acceptUntil(tok tokenType) (t []token) {
	for {
		curr := p.next()
		switch curr.typ {
		case tok:
			p.backup()
			return
		case tokenEOF, tokenError: //eof and error signify no more tokens
			return
		}
		t = append(t, curr)
	}
	panic("unreachable")
}

func subParse(parp *parser, end tokenType) (ex executer, err error) {
	p := &parser{
		in:   parp.in,
		out:  make(chan executer),
		err:  make(chan error, 1),
		end:  end,
		errd: tokenNone,
	}
	go p.run()

	//grab our executers
	l := executeList{}
	for e := range p.out {
		l.Push(e)
	}

	ex = l

	//grab an error if it happened
	select {
	case err = <-p.err:
	default:
	}

	parp.curr = p.curr
	parp.backed = p.backed
	parp.errd = p.errd

	return
}

func parseText(p *parser) (s parseState) {
	//only accept literal, open, and eof
	switch tok := p.next(); tok.typ {
	case tokenLiteral:
		p.out <- constantValue(tok.dat)
		return parseText
	case tokenOpen:
		return parseOpen
	case tokenEOF:
		if p.end == tokenNoneType {
			return nil
		}
		return p.errorf("unexpected eof. in a %q context", p.end)
	default:
		return p.errorf("Unexpected token: %s", tok)
	}
	return nil
}

func parseOpen(p *parser) parseState {
	switch tok := p.next(); {
	//advanced calls to start a sub parser
	case tok.typ == tokenBlock:
		return parseBlock
	case tok.typ == tokenWith:
		return parseWith
	case tok.typ == tokenRange:
		return parseRange
	case tok.typ == tokenIf:
		return parseIf

	//very special call to handle else
	case tok.typ == tokenElse:
		if p.end != tokenIf {
			return p.errorf("Unexpected else not inside an if context")
		}
		return nil

	//value calls
	case isValueType(tok):
		p.backup()
		val, s := consumeValue(p)
		if s != nil {
			return p.errorf(s.Error())
		}

		//grab the close
		if t := p.next(); t.typ != tokenClose {
			return p.errExpect(tokenClose, t)
		}

		p.out <- val
		return parseText

	//end tag
	case tok.typ == tokenEnd:
		return parseEnd

	//do nothing
	case tok.typ == tokenClose:
		return parseText
	default:
		return p.unexpected(tok)
	}
	panic("unreachable")
}

//parse end should signal the end of a sub parser
func parseEnd(p *parser) parseState {
	//didn't get the end we're looking for
	if tok := p.next(); tok.typ != p.end {
		return p.errExpect(p.end, tok)
	}
	if tok := p.next(); tok.typ != tokenClose {
		return p.errExpect(tokenClose, tok)
	}
	return nil
}

func parseBlock(p *parser) parseState {
	//grab the name
	ident := p.next()
	if ident.typ != tokenIdent {
		return p.errExpect(tokenIdent, ident)
	}

	//see if we have a value type
	var ctx valueType
	if isValueType(p.peek()) {
		var err error
		ctx, err = consumeValue(p)
		if err != nil {
			return p.errorf(err.Error())
		}
	}

	//consume the close
	if tok := p.next(); tok.typ != tokenClose {
		return p.errExpect(tokenClose, tok)
	}

	//start a sub parser looking for an end block
	ex, err := subParse(p, tokenBlock)
	if err != nil {
		return p.errorf(err.Error())
	}

	p.out <- &executeBlock{string(ident.dat), ctx, ex}
	return parseText
}

func parseWith(p *parser) parseState {
	//grab the value type
	ctx, st := consumeValue(p)
	if st != nil {
		return p.errorf(st.Error())
	}

	//grab the close
	if tok := p.next(); tok.typ != tokenClose {
		return p.errExpect(tokenClose, tok)
	}

	ex, err := subParse(p, tokenWith)
	if err != nil {
		return p.errorf(err.Error())
	}

	p.out <- &executeWith{ctx, ex}
	return parseText
}

func parseRange(p *parser) parseState {
	//grab the value type
	ctx, st := consumeValue(p)
	if st != nil {
		return p.errorf(st.Error())
	}

	//grab the close
	if tok := p.next(); tok.typ != tokenClose {
		return p.errExpect(tokenClose, tok)
	}

	ex, err := subParse(p, tokenRange)
	if err != nil {
		return p.errorf(err.Error())
	}

	p.out <- &executeRange{ctx, ex}
	return parseText
}

func parseIf(p *parser) parseState {
	//grab the value
	cond, st := consumeValue(p)
	if st != nil {
		return p.errorf(st.Error())
	}

	//grab the close
	if tok := p.next(); tok.typ != tokenClose {
		return p.errExpect(tokenClose, tok)
	}

	//start a sub parser for succ
	succ, err := subParse(p, tokenIf)
	if err != nil {
		return p.errorf(err.Error())
	}

	//backup to check how we exited
	p.backup()

	var fail executer
	switch tok := p.next(); tok.typ {
	case tokenElse:
		//grab the close
		if tok := p.next(); tok.typ != tokenClose {
			return p.errExpect(tokenClose, tok)
		}

		var err error
		fail, err = subParse(p, tokenIf)
		if err != nil {
			return p.errorf(err.Error())
		}
	case tokenClose:
	default:
		return p.unexpected(tok)
	}

	p.out <- &executeIf{cond, succ, fail}
	return parseText
}
