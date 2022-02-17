package org

import (
	"strings"
	"unicode/utf8"
)

const (
	eof rune = 0
)

type itemType int

const (
	lexEmpty itemType = iota
	lexTextLine
	lexCommentBody
	lexCommentStart
	lexKeywordLine
	lexHeadingLevel
	lexHeadingBody
)

type item struct {
	typ   itemType
	text  string
	start int
	end   int
}

type lexer struct {
	name  string
	input string
	start int
	pos   int
	width int
	items chan item
}

type stateFn func(l *lexer) stateFn

func (l *lexer) emit(i item) {
	l.items <- i
}

func (l *lexer) close() {
	close(l.items)
}

func (l *lexer) next() (r rune) {
	if l.pos >= len(l.input) {
		l.width = 0
		return eof
	}
	r, l.width = utf8.DecodeLastRuneInString(l.input[l.pos:])
	l.pos += l.width
	return
}

func (l *lexer) isEOF() bool {
	return l.pos >= len(l.input)
}

func (l *lexer) backup() {
	l.pos -= l.width
	l.width = 0
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) peek() (r rune) {
	if l.pos >= len(l.input) {
		return eof
	}
	r, _ = utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

func (l *lexer) accept(valid string) bool {
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) acceptWhile(accptFn func(rune) bool) {
	for accptFn(l.next()) {
	}
	if !l.isEOF() {
		l.backup()
	}
}

func (l *lexer) acceptUntil(untilFn func(rune) bool) {
	for r := l.next(); r != eof && !untilFn(r); r = l.next() {
	}
	if !l.isEOF() {
		l.backup()
	}
}

func (l *lexer) acceptUntilEOL() {
	for r := l.next(); r != eof && r != '\n'; r = l.next() {
	}
	l.backup()
}
