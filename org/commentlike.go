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

func tryHeading(start int, l *lexer) int {
	l.acceptRun("*")
	if strings.IndexRune(" \t", l.next()) >= 0 {
		return l.pos - start - 1
	}
	return 0
}

func tryDrawer(l *lexer) bool {
	l.acceptRun("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-_$%'")
	if l.next() == ':' {
		return true
	}
	return false
}

func lexDefault(l *lexer) stateFn {
	pos := l.pos
	switch l.next() {
	case eof:
		l.close()
		return nil
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
			return lexComment(pos)
		case keywordStart:
			return lexKeyword
		}
	case ':':
		if tryDrawer(l) {
			return lexDrawer
		}
	case '*':
		level := tryHeading(pos, l)
		if level > 0 {
			return lexHeading(pos, level)
		}
	case ' ', '\t':
		return lexNotHeading
	}
	return lexText
}

func lexNotHeading(l *lexer) stateFn {
	pos := l.pos
	switch l.next() {
	case eof:
		l.close()
		return nil
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
			return lexComment(pos)
		case keywordStart:
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

func lexHeading(start, level int) func(*lexer) stateFn {
	return func(l *lexer) stateFn {
		l.emit(item{
			typ:   lexHeadingLevel,
			text:  l.input[start : start+level],
			start: start,
			end:   start + level,
		})
		l.acceptUntilEOL()
		l.emit(item{
			typ:   lexHeadingBody,
			text:  l.input[l.start : l.pos-1],
			start: l.start,
			end:   l.pos,
		})
		l.start = l.pos
		return lexDefault
	}
}

func lexComment(start int) func(*lexer) stateFn {
	return func(l *lexer) stateFn {
		l.emit(item{
			typ:   lexCommentStart,
			text:  l.input[start : start+1],
			start: start,
			end:   start,
		})
		l.acceptUntilEOL()
		l.emit(item{
			typ:   lexCommentBody,
			text:  l.input[l.start : l.pos-1],
			start: l.start,
			end:   l.pos,
		})
		l.start = l.pos
		return lexDefault
	}
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
