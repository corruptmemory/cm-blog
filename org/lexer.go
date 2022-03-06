package org

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

const eof = -1

type itemType int

const (
	itemError itemType = iota
	itemLeadingSpace
	itemNewline
	itemEOF
	itemTextLine
	itemComment
	itemKeyword
	itemHeading
	itemNodeMarker
)

// Pos represents a byte position in the original input text from which
// this template was parsed.
type Pos int

func (p Pos) Position() Pos {
	return p
}

type item struct {
	typ  itemType
	pos  Pos
	val  string
	line int
}

type lexer struct {
	name      string
	input     string
	start     Pos
	pos       Pos
	width     Pos
	items     chan item
	line      int
	startLine int
}

type stateFn func(l *lexer) stateFn

// lex creates a new scanner for the input string.
func lex(name, input string) *lexer {
	l := &lexer{
		name:      name,
		input:     input,
		items:     make(chan item),
		line:      1,
		startLine: 1,
	}
	go l.run()
	return l
}

func (l *lexer) emit(t itemType) {
	l.items <- item{t, l.start, l.input[l.start:l.pos], l.startLine}
	l.start = l.pos
	l.startLine = l.line
}

func (l *lexer) close() {
	close(l.items)
}

func (l *lexer) next() (r rune) {
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = Pos(w)
	l.pos += l.width
	if r == '\n' {
		l.line++
	}
	return r
}

func (l *lexer) isEOF() bool {
	return int(l.pos) >= len(l.input)
}

func (l *lexer) backup() {
	l.pos -= l.width
	// Correct newline count.
	if l.width == 1 && l.input[l.pos] == '\n' {
		l.line--
	}
}

func (l *lexer) ignore() {
	l.line += strings.Count(l.input[l.start:l.pos], "\n")
	l.start = l.pos
	l.startLine = l.line
}

func (l *lexer) run() {
	for lex := lexDefault; lex != nil; {
		lex = lex(l)
	}
}

func (l *lexer) peek() (r rune) {
	if int(l.pos) >= len(l.input) {
		return eof
	}
	r, _ = utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.ContainsRune(valid, l.next()) {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.ContainsRune(valid, l.next()) {
	}
	l.backup()
}

func (l *lexer) acceptWhile(accptFn func(rune) bool) {
	for r := l.next(); r != eof && accptFn(r); r = l.next() {
	}
	l.backup()
}

func (l *lexer) acceptUntil(untilFn func(rune) bool) {
	l.acceptWhile(func(r rune) bool {
		return !untilFn(r)
	})
}

func (l *lexer) acceptUntilEOL() {
	l.acceptUntil(func(r rune) bool {
		return r == '\n'
	})
}

func (l *lexer) ignoreToEOL() {
	idx := strings.IndexByte(l.input[l.pos:], '\n')
	if idx >= 0 {
		l.pos = l.pos + Pos(idx+1)
	} else {
		l.pos = Pos(len(l.input))
	}
	l.ignore()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...), l.startLine}
	return nil
}

func lexDefault(l *lexer) stateFn {
	switch l.next() {
	case eof:
		l.emit(itemEOF)
		l.close()
		return nil
	case ' ', '\t':
		return lexSpace
	case '\n':
		l.emit(itemNewline)
		return lexDefault
	case '#':
		return lexCommentOrKeyword
	case ':':
		return lexDrawer
	case '*':
		return lexHeading
	}
	return lexText
}

func lexAfterLeadingSpace(l *lexer) stateFn {
	switch l.next() {
	case eof:
		l.emit(itemEOF)
		l.close()
		return nil
	case '\n':
		l.emit(itemNewline)
		return lexDefault
	case '#':
		return lexCommentOrKeyword
	case ':':
		return lexDrawer
	}
	return lexText
}

func lexSpace(l *lexer) stateFn {
	l.acceptRun(" \t")
	l.emit(itemLeadingSpace)
	return lexAfterLeadingSpace
}

func lexHeading(l *lexer) stateFn {
	l.acceptRun("*")
	switch l.peek() {
	case ' ', '\t':
		l.acceptUntilEOL()
		l.emit(itemHeading)
		return lexDefault
	}
	return lexText
}

func lexCommentOrKeyword(l *lexer) stateFn {
	switch l.peek() {
	case ' ', '\t':
		return lexComment
	case '+':
		return lexKeyword
	}
	return lexText
}

func lexComment(l *lexer) stateFn {
	switch l.peek() {
	case ' ', '\t':
		l.acceptUntilEOL()
		l.emit(itemComment)
		return lexDefault
	}
	return lexText
}

func lexKeyword(l *lexer) stateFn {
	switch l.peek() {
	case '+':
		l.accept("+")
		if l.accept(" \t") {
			return lexText
		}
		l.acceptUntilEOL()
		l.emit(itemKeyword)
		return lexDefault
	}
	return lexText
}

func lexPossibleNodeMarker(l *lexer) stateFn {
	switch l.peek() {
	case eof, '\n':
		l.emit(itemNodeMarker)
	case ' ', '\t':
		l.acceptUntilEOL()
		l.emit(itemNodeMarker)
	default:
		return lexText
	}
	return lexDefault
}

func lexDrawer(l *lexer) stateFn {
	l.acceptRun("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_$%'+")
	if l.next() == ':' {
		return lexPossibleNodeMarker
	}
	return lexText
}

func lexText(l *lexer) stateFn {
	l.acceptUntilEOL()
	l.emit(itemTextLine)
	return lexDefault
}
