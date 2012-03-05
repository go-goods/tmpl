package tmpl

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenType int

//TODO: unicode support
const identifierLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_0123456789"

const (
	tokenOpen     tokenType = iota // {%
	tokenClose                     // %}
	tokenCall                      // call
	tokenPush                      // .
	tokenPop                       // $
	tokenRoot                      // /
	tokenValue                     // "foo"
	tokenNumeric                   // -123.5
	tokenIdent                     // foo (push/pop idents)
	tokenAs                        // as
	tokenBlock                     // block
	tokenEvoke                     // evoke
	tokenIf                        // if
	tokenElse                      // else
	tokenWith                      // with
	tokenRange                     // range
	tokenEnd                       // end
	tokenComment                   // comment
	tokenLiteral                   // stuff between open/close
	tokenEOF                       // sent when no data is left
	tokenStartSel                  // sent at the start of a selector like .foo$bar
	tokenEndSel                    // sent at the end of a select like .foo$bar
	tokenError                     // error type

	//special sentinal value used in the parser
	tokenNoneType tokenType = -1
)

var (
	commentOpen  = []byte(`{#`)
	commentClose = []byte(`#}`)
)

var tokenNames = []string{
	"open", "close", "call", "push", "pop", "root", "value", "numeric", "ident",
	"as", "block", "evoke", "if", "else", "with", "range", "end", "comment",
	"literal", "eof", "startSel", "endSel", "error",
}

func (t tokenType) String() string {
	if t == -1 {
		return "NONE"
	}
	return tokenNames[t]
}

func isErrorType(t tokenType) bool {
	switch t {
	case tokenEOF, tokenError:
		return true
	}
	return false
}

const eof rune = -1

type delim struct {
	value []byte
	typ   tokenType
}

var (
	openDelim  = delim{[]byte(`{%`), tokenOpen}
	closeDelim = delim{[]byte(`%}`), tokenClose}
	pushDelim  = delim{[]byte(`.`), tokenPush}
	popDelim   = delim{[]byte(`$`), tokenPop}
	rootDelim  = delim{[]byte(`/`), tokenRoot}
	callDelim  = delim{[]byte(`call`), tokenCall}
	blockDelim = delim{[]byte(`block`), tokenBlock}
	evokeDelim = delim{[]byte(`evoke`), tokenEvoke}
	ifDelim    = delim{[]byte(`if`), tokenIf}
	elseDelim  = delim{[]byte(`else`), tokenElse}
	withDelim  = delim{[]byte(`with`), tokenWith}
	rangeDelim = delim{[]byte(`range`), tokenRange}
	asDelim    = delim{[]byte(`as`), tokenAs}
	endDelim   = delim{[]byte(`end`), tokenEnd}

	insideDelims = []delim{callDelim, blockDelim, ifDelim, elseDelim, withDelim, rangeDelim, endDelim, asDelim, evokeDelim}
	selDelims    = []delim{pushDelim, popDelim, rootDelim}
)

type token struct {
	typ       tokenType
	dat       []byte
	line, pos int
}

var tokenNone = token{tokenNoneType, nil, 0, 0}

func (t token) String() string {
	return fmt.Sprintf("%d:%d[%s]%s", t.line, t.pos, tokenNames[t.typ], t.dat)
}

type lexer struct {
	data   []byte
	pos    int
	lines  int
	lastnl int
	tail   int
	width  int
	pipe   chan token
}

type lexerState func(l *lexer) lexerState

func lex(data []byte) chan token {
	l := &lexer{
		data: data,
		pipe: make(chan token),
	}
	go l.run()
	return l.pipe
}

//run runs the state machine
func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.pipe)
}

//slice returns the current token value
func (l *lexer) slice() []byte {
	return l.data[l.tail:l.pos]
}

//advance moves the tail up to the pos igoring the current token
func (l *lexer) advance() {
	l.tail = l.pos
}

//next advances the post token one rune
func (l *lexer) next() (r rune) {
	if l.pos >= len(l.data) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRune(l.data[l.pos:])
	l.pos += l.width
	return
}

//backup backs up the last rune returned by next
func (l *lexer) backup() {
	l.pos -= l.width
	l.width = 0
}

//emit sends out the current token with the given type
func (l *lexer) emit(typ tokenType) {
	//figure out how many more newlines to add
	dat := l.slice()
	l.pipe <- token{
		typ:  typ,
		dat:  dat,
		pos:  l.tail - l.lastnl,
		line: l.lines,
	}
	newlines := bytes.Count(dat, []byte{'\n'})
	l.lines += newlines
	if newlines > 0 {
		l.lastnl = l.tail + bytes.LastIndex(dat, []byte{'\n'}) + 1
	}
	l.advance()
}

//accept takes a set of valid chars and accepts the next character if it is
//in the set. returns if the character was accepted.
func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

//acceptRun accepts like the regex [valid]*
func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) acceptUntil(invalid string) {
	for strings.IndexRune(invalid, l.next()) == -1 {
	}
	l.backup()
}

//peek returns the next rune without moving the pointer
func (l *lexer) peek() (r rune) {
	r = l.next()
	l.backup()
	return
}

func (l *lexer) errorf(format string, args ...interface{}) lexerState {
	l.pipe <- token{
		typ:  tokenError,
		dat:  []byte(fmt.Sprintf(format, args...)),
		pos:  l.tail - l.lastnl,
		line: l.lines,
	}
	return nil
}

func lexText(l *lexer) lexerState {
	for {
		//open tags
		if bytes.HasPrefix(l.data[l.pos:], openDelim.value) {
			//check if we should emit
			if l.pos > l.tail {
				l.emit(tokenLiteral)
			}
			return lexOpenDelim
		}

		//comments
		if bytes.HasPrefix(l.data[l.pos:], commentOpen) {
			//check if we should emit
			if l.pos > l.tail {
				l.emit(tokenLiteral)
			}
			return lexComment
		}

		//check for eof
		if l.next() == eof {
			break
		}
	}

	//correctly reached an eof
	if l.pos > l.tail {
		l.emit(tokenLiteral)
	}

	//send an eof
	l.emit(tokenEOF)
	return nil
}

func lexOpenDelim(l *lexer) lexerState {
	l.pos += len(openDelim.value)
	l.emit(openDelim.typ)
	return lexInsideDelims
}

func lexCloseDelim(l *lexer) lexerState {
	l.pos += len(closeDelim.value)
	l.emit(closeDelim.typ)
	return lexText
}

func lexPushDelim(l *lexer) lexerState {
	l.pos += len(pushDelim.value)
	l.emit(pushDelim.typ)
	return lexInsideSel
}

func lexPopDelim(l *lexer) lexerState {
	l.pos += len(popDelim.value)
	l.emit(popDelim.typ)
	return lexInsideSel
}

func lexInsideDelims(l *lexer) lexerState {
	for {
		rest := l.data[l.pos:]
		//lex the inside tokens that dont change state
		for _, delim := range insideDelims {
			if bytes.HasPrefix(rest, delim.value) {
				l.pos += len(delim.value)
				l.emit(delim.typ)
				return lexInsideDelims
			}
		}

		//check for things that start selectors
		for _, delim := range selDelims {
			if bytes.HasPrefix(rest, delim.value) {
				l.emit(tokenStartSel)
				return lexInsideSel
			}
		}

		//check for a close delim
		if bytes.HasPrefix(rest, closeDelim.value) {
			return lexCloseDelim
		}

		switch r := l.next(); {
		case r == eof || r == '\n' || r == '\r':
			return l.errorf("unclosed action")
		case unicode.IsSpace(r):
			l.advance()
		case r == '+' || r == '-' || '0' <= r && r <= '9':
			l.backup()
			return lexNumber
		case r == '"':
			l.advance()
			return lexValue
		case unicode.IsLetter(r) || r == '_': //go spec
			return lexIdentifier
		default:
			return l.errorf("invalid character: %q", r)
		}
	}
	return nil
}

func lexComment(l *lexer) lexerState {
	for !bytes.HasPrefix(l.data[l.pos:], commentClose) {
		if l.next() == eof {
			return l.errorf("unexpected eof in comment")
		}
	}
	l.pos += len(commentClose)
	l.emit(tokenComment)
	return lexText
}

func lexInsideSel(l *lexer) lexerState {
	for {
		rest := l.data[l.pos:]
		if bytes.HasPrefix(rest, pushDelim.value) {
			return lexPushDelim
		}
		if bytes.HasPrefix(rest, popDelim.value) {
			return lexPopDelim
		}
		switch r := l.next(); {
		case unicode.IsLetter(r) || r == '_': //go spec
			l.acceptRun(identifierLetters)
			l.emit(tokenIdent)
			return lexInsideSel
		case unicode.IsSpace(r):
			l.emit(tokenEndSel)
			return lexInsideDelims
		default:
			return l.errorf("invalid character: %q", r)
		}
	}
	return nil
}

func lexValue(l *lexer) lexerState {
	l.acceptUntil(`"`)
	l.emit(tokenValue)
	l.next() //grab the right quote and chunk it
	l.advance()
	return lexInsideDelims
}

func lexIdentifier(l *lexer) lexerState {
	l.acceptRun(identifierLetters)
	l.emit(tokenIdent)
	return lexInsideDelims
}

func lexNumber(l *lexer) lexerState {
	//optional leading sign
	l.accept("+-")
	l.acceptRun("0123456789")
	if l.accept(".") {
		l.acceptRun("0123456789")
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	l.emit(tokenNumeric)
	return lexInsideDelims
}
