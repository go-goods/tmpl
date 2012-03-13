package tmpl

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

//parseTree represents a parsed template and the set of blocks/functions that
//it will use to execute.
type parseTree struct {
	base    executer
	context *context
}

//Execute runs the parsed template with the context value as the root.
func (p *parseTree) Execute(w io.Writer, ctx interface{}) error {
	if p.base == nil {
		return nil
	}
	p.context.stack = pathRootedAt(ctx)
	return p.base.Execute(w, p.context)
}

//String returns a nice printable representation of the parse tree and context.
func (p *parseTree) String() string {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, p.context)
	fmt.Fprintln(&buf, p.base)
	return buf.String()
}

//parser is a type that represnts an ongoing parse of a template.
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

//parseState is a transition state of the parser state machine.
type parseState func(*parser) parseState

//parse compiles the incoming channel of tokens into a parseTree.
func parse(toks chan token) (t *parseTree, err error) {
	t = &parseTree{
		context: newContext(),
	}
	//make a channel of blocks to stick into the context
	blocks := make(chan *executeBlockValue)
	go func() {
		//start a new parser
		p := &parser{
			in:     toks,
			errd:   tokenNone,
			blocks: blocks,
		}
		//set the base of the parse tree
		t.base, err = subParse(p, tokenNoneType)
		//signal no more blocks are coming
		close(blocks)
	}()

	//array for redefined block errors
	var redef []string
	for b := range blocks {
		//check if we're redefining a block
		if _, ex := t.context.blocks[b.ident]; ex {
			redef = append(redef, fmt.Sprintf("Redefined block %s", b.ident))
		}
		//set our block
		t.context.blocks[b.ident] = b
	}

	//return an error about redefined blocks
	if redef != nil {
		err = fmt.Errorf(strings.Join(redef, "\n"))
	}

	//if we have an error, don't return a parse tree
	if err != nil {
		t = nil
	}

	return
}

//run executes the parser state machine
func (p *parser) run() {
	for state := parseText; state != nil; {
		state = state(p)
	}
	close(p.out)
}

//errorf is a helper that sets an error and returns a stop state
func (p *parser) errorf(format string, args ...interface{}) parseState {
	p.err = fmt.Errorf(format, args...)
	return nil
}

//errExpect is a helper that sets an error and returns a stop state
func (p *parser) errExpect(ex tokenType, got token) parseState {
	return p.errorf("Compile: Expected a %q got a %q", ex, got)
}

//unexpected is a helper that sets an error and returns a stop state
func (p *parser) unexpected(t token) parseState {
	return p.errorf("Unexpected %q", t)
}

//accept will accept a token of the given type, and return if it did
func (p *parser) accept(tok tokenType) bool {
	if p.next().typ == tok {
		return true
	}
	p.backup()
	return false
}

//next returns the next token from the channel. If an error is ever encountered,
//it will return that token every time. It also respects backing up.
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

//backup puts the last read token back into the channel to be read again. It
//will panic if backup happens more than once.
func (p *parser) backup() {
	if p.backed {
		panic("double backup")
	}
	p.backed = true
}

//peek returns the next token in the channel without consuming it. It is
//equivelant to a next() and a backup()
func (p *parser) peek() (t token) {
	t = p.next()
	p.backup()
	return
}

//acceptUntil accepts tokens until a token of the given type is found and it
//will backup so that the next token is of the given type. It will also stop
//if an error type is encountered.
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

//subParse starts another parser that runs until an end clause is encountered
//of the given tokenType.
func subParse(parp *parser, end tokenType) (ex executer, err error) {
	//create our sub-parser
	p := &parser{
		in:      parp.in,                           //use the same in channel
		out:     make(chan executer),               //make a new out channel
		end:     end,                               //look for the given end token
		errd:    tokenNone,                         //we haven't errored yet
		inBlock: parp.inBlock || end == tokenBlock, //check if we're in a block
		blocks:  parp.blocks,                       //use the same block channel
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

//parseText is the start state of the parser.
func parseText(p *parser) (s parseState) {
	switch tok := p.next(); tok.typ {
	//do nothing with comments
	case tokenComment:
		return parseText
	//send out literal values and keep parsing text
	case tokenLiteral:
		p.out <- constantValue(tok.dat)
		return parseText
	//with an open check what the next action is
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

//parseOpen is the state after an open action token.
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

//parseEnd should signal the end of a sub parser
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

//parseEvoke parses an evoke action.
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

//parseBlock parses a block definition.
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

	//send it to blocks instead of out
	p.blocks <- &executeBlockValue{string(ident.dat), "", ex}
	return parseText
}

//parseWith parses a with action.
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

//parseRange parses a range action.
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

//parseIf parses an if clause.
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
