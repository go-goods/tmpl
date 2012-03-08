package tmpl

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

type parseTree struct {
	base    executer
	context *context
}

func (p *parseTree) Execute(w io.Writer, ctx interface{}) error {
	if p.base == nil {
		return nil
	}
	p.context.stack = pathRootedAt(ctx)
	return p.base.Execute(w, p.context)
}

func (p *parseTree) String() string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, p.context)
	fmt.Fprintln(&buf, p.base)
	return buf.String()
}

type parser struct {
	//parser setup
	in  chan token    //token channel
	out chan executer //output channel
	err error         //error during parsing
	end tokenType     //for subparse to check for the correct end type

	//block channel
	blocks  chan *executeBlockValue
	inBlock bool

	//token state types
	curr   token //currently read token
	backed bool  //if we're in a backup state
	errd   token //if a token is an EOF or Error to repeat it forever
}

type parseState func(*parser) parseState

func parse(toks chan token) (t *parseTree, err error) {
	t = &parseTree{
		context: newContext(),
	}
	blocks := make(chan *executeBlockValue)
	go func() {
		p := &parser{
			in:     toks,
			errd:   tokenNone,
			blocks: blocks,
		}
		t.base, err = subParse(p, tokenNoneType)
		close(blocks)
	}()

	var redef []string
	for b := range blocks {
		if _, ex := t.context.blocks[b.ident]; ex {
			redef = append(redef, fmt.Sprintf("Redefined block %s", b.ident))
		}
		t.context.blocks[b.ident] = b
	}

	if redef != nil {
		err = fmt.Errorf(strings.Join(redef, "\n"))
	}

	if err != nil {
		t = nil
	}

	return
}

func (p *parser) run() {
	for state := parseText; state != nil; {
		state = state(p)
	}
	close(p.out)
}

func (p *parser) errorf(format string, args ...interface{}) parseState {
	p.err = fmt.Errorf(format, args...)
	return nil
}

func (p *parser) errExpect(ex tokenType, got token) parseState {
	return p.errorf("Compile: Expected a %q got a %q", ex, got)
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
	if isErrorType(p.curr.typ) {
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
		switch typ := curr.typ; {
		case typ == tok:
			p.backup()
			return
		case isErrorType(typ): //eof and error signify no more tokens
			return
		}
		t = append(t, curr)
	}
	panic("unreachable")
}

func subParse(parp *parser, end tokenType) (ex executer, err error) {
	p := &parser{
		in:      parp.in,
		out:     make(chan executer),
		end:     end,
		errd:    tokenNone,
		inBlock: parp.inBlock || end == tokenBlock,
		blocks:  parp.blocks,
	}
	//run the parser
	go p.run()

	//grab our executers
	l := executeList{}
	for e := range p.out {
		l.Push(e)
	}
	//compact the list for execute efficiency
	l.compact()

	//set our executer, dropping the list if it is one element
	switch len(l) {
	case 0:
	case 1:
		ex = l[0]
	default:
		ex = l
	}

	//grab an error if it happened
	err = p.err

	//set the token state on the parent to make backup/peek work
	parp.curr = p.curr
	parp.backed = p.backed
	parp.errd = p.errd

	return
}

func parseText(p *parser) (s parseState) {
	//only accept literal, open, and eof
	switch tok := p.next(); tok.typ {
	case tokenComment:
		return parseText
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
		//check for a sub parse
		if p.inBlock {
			return p.errorf("%d:%d: nested blocks", tok.line, tok.pos)
		}
		return parseBlock
	case tok.typ == tokenWith:
		return parseWith
	case tok.typ == tokenRange:
		return parseRange
	case tok.typ == tokenIf:
		return parseIf
	case tok.typ == tokenEvoke:
		return parseEvoke

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

func parseEvoke(p *parser) parseState {
	//grab the name
	ident := p.next()
	if ident.typ != tokenIdent {
		return p.errExpect(tokenIdent, ident)
	}

	//see if we have a value type
	var ctx *selectorValue
	if isValueType(p.peek()) {
		var err error
		ctx, err = consumeSelector(p)
		if err != nil {
			return p.errorf(err.Error())
		}
	}

	//grab the close
	if tok := p.next(); tok.typ != tokenClose {
		return p.errExpect(tokenClose, tok)
	}

	p.out <- &executeEvoke{string(ident.dat), ctx}
	return parseText
}

func parseBlock(p *parser) parseState {
	//grab the name
	ident := p.next()
	if ident.typ != tokenIdent {
		return p.errExpect(tokenIdent, ident)
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

	p.blocks <- &executeBlockValue{string(ident.dat), "", ex}
	return parseText
}

func parseWith(p *parser) parseState {
	//grab the value type
	ctx, st := consumeSelector(p)
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

	//default to none
	key, val := tokenNone, tokenNone

	//check for an as
	if tok := p.next(); tok.typ == tokenAs {
		//grab two idenitifers
		if key = p.next(); key.typ != tokenIdent {
			return p.errExpect(tokenIdent, key)
		}
		if val = p.next(); val.typ != tokenIdent {
			return p.errExpect(tokenIdent, val)
		}
	} else {
		//whoops wasn't an as
		p.backup()
	}

	//grab the close
	if tok := p.next(); tok.typ != tokenClose {
		return p.errExpect(tokenClose, tok)
	}

	ex, err := subParse(p, tokenRange)
	if err != nil {
		return p.errorf(err.Error())
	}

	p.out <- &executeRange{ctx, ex, key, val}
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
