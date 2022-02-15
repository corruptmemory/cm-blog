package org

import (
	"strings"
)

type commentResult int

const (
	commentContinueText commentResult = iota
	commentStart
	keywordStart
)

func tryComment(l *lexer) commentResult {
	var t commentResult
	r := l.next()
	switch {
	case strings.IndexRune(" \t", r) >= 0:
		t = commentStart
	case r == '+':
		t = keywordStart
	}
	return t
}

func tryHeading(l *lexer) bool {
	l.acceptRun("*")
	if strings.IndexRune(" \t", l.next()) >= 0 {
		return true
	}
	return false
}

func tryDrawer(l *lexer) bool {
	l.acceptRun("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_$%'")
	if l.next() == ':' {
		return true
	}
	return false
}

func lexDefault(l *lexer) stateFn {
	switch l.next() {
	case '\n':
		l.emit(item{
			typ:   lexTextLine,
			text:  l.input[l.start : l.pos-1],
			start: l.start,
			end:   l.pos,
		})
		l.start = l.pos
		return lexDefault
	case '#':
		switch tryComment(l) {
		case commentStart:
			l.ignore()
			return lexComment
		case keywordStart:
			l.ignore()
			return lexKeyword
		}
	case ':':
		if tryDrawer(l) {
			return lexDrawer
		}
	case '*':
		if tryHeading(l) {
			return lexHeading
		}
	case ' ', '\t':
		return lexNotHeading
	}
	return lexText
}

func lexNotHeading(l *lexer) stateFn {
	switch l.next() {
	case '\n':
		l.emit(item{
			typ:   lexTextLine,
			text:  l.input[l.start : l.pos-1],
			start: l.start,
			end:   l.pos,
		})
		l.start = l.pos
		return lexDefault
	case '#':
		switch tryComment(l) {
		case commentStart:
			l.ignore()
			return lexComment
		case keywordStart:
			l.ignore()
			return lexKeyword
		}
	case ':':
		if tryDrawer(l) {
			return lexDrawer
		}
	case ' ', '\t':
		return lexNotHeading
	}
	return lexText
}

func lexHeading(l *lexer) stateFn {
	l.acceptUntilEOL()
	l.emit(item{
		typ:   lexHeadingLine,
		text:  l.input[l.start : l.pos-1],
		start: l.start,
		end:   l.pos,
	})
	l.start = l.pos
	return lexDefault
}

func lexComment(l *lexer) stateFn {
	l.acceptUntilEOL()
	l.emit(item{
		typ:   lexCommentLine,
		text:  l.input[l.start : l.pos-1],
		start: l.start,
		end:   l.pos,
	})
	l.start = l.pos
	return lexDefault
}

func lexKeyword(l *lexer) stateFn {
	return nil
}

func lexDrawer(l *lexer) stateFn {
	return nil
}

func lexText(l *lexer) stateFn {
	return nil
}
