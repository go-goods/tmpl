package tmpl

import (
	"bytes"
	"unicode/utf8"
)

type tokenType int

const (
	tokenOpen    tokenType = iota // {%
	tokenClose                    // %}
	tokenCall                     // call
	tokenPush                     // .
	tokenPop                      // $
	tokenValue                    // "foo"
	tokenNumeric                  // -123.5
	tokenIdent                    // foo (push/pop idents)
	tokenBlock                    // block
	tokenIf                       // if
	tokenElse                     // else
	tokenWith                     // with
	tokenRange                    // range
	tokenEnd                      // end
	tokenLiteral                  // stuff between open/close
	tokenEOF                      // sent when no data is left
	tokenError                    // error type
)

const eof rune = -1

var (
	openDelim  = []byte(`{%`)
	closeDelim = []byte(`%}`)
	pushDelim  = []byte(`.`)
	popDelim   = []byte(`$`)
	callDelim  = []byte(`call`)
)

type token struct {
	typ tokenType
	dat []byte
}

type lexer struct {
	data  []byte
	pos   int
	tail  int
	width int
	pipe  chan token
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

func (l *lexer) run() {
	for state := lexText; state != nil; {
		state = state(l)
	}
	close(l.pipe)
}

func (l *lexer) slice() []byte {
	return l.data[l.tail:l.pos]
}

func (l *lexer) advance() {
	l.tail = l.pos
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.data) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeRune(l.data[l.pos:])
	l.pos += l.width
	return
}

func (l *lexer) backup() {
	l.pos -= l.width
	l.width = 0
}

func (l *lexer) emit(typ tokenType) {
	l.pipe <- token{
		typ: typ,
		dat: l.slice(),
	}
	l.advance()
}

func (l *lexer) accept(valid []byte) bool {
	if bytes.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid []byte) {
	for bytes.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) peek() (r rune) {
	r = l.next()
	l.backup()
	return
}

func lexText(l *lexer) lexerState {
	for {
		if bytes.HasPrefix(l.data[l.pos:], openDelim) {
			//check if we should emit
			if l.pos > l.tail {
				l.emit(tokenLiteral)
			}
			return lexOpenDelim
		}
		if l.next() == eof {
			break
		}
	}
	//correctly reached an eof
	if l.pos > l.tail {
		l.emit(tokenLiteral)
	}
	l.emit(tokenEOF)
	return nil
}

func lexOpenDelim(l *lexer) lexerState {
	l.pos += len(openDelim)
	l.emit(tokenOpen)
	return lexInsideDelims
}

func lexCloseDelim(l *lexer) lexerState {
	l.pos += len(closeDelim)
	l.emit(tokenClose)
	return lexText
}

func lexInsideDelims(l *lexer) lexerState {
	for {
		if bytes.HasPrefix(l.data[l.pos:], closeDelim) {
			return lexCloseDelim
		}
		switch r := l.next(); {
		case true:
			_ = r
		}
	}
}

func lexNumber(l *lexer) lexerState {
	//optional leading sing
	l.accept("+-")
	//is it hex?
	digits := "0123456789"
	if l.accept("0") && l.accept("xX") {
		digits = "0123456789abcdefABCDEF"
	}
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun("0123456789")
	}
	l.emit(tokenNumeric)
	return lexInsideDelims
}
