package org

import (
	"strings"
)

func tryComment(l *lexer) itemType {
	commentStart := l.pos
	if l.accept("#") {
		r := l.next()
		var t itemType
		switch {
		case strings.IndexRune(" \t", r) >= 0:
			t = lexCommentStart
		case r == '+':
			t = lexKeywordStart
		}
		if t != lexEmpty {
			l.emit(item{
				typ:   t,
				text:  l.input[commentStart:l.pos],
				start: commentStart,
				end:   l.pos,
			})
			return lexDefault
		}
	}
	l.pos = commentStart
	return fallback
}

func lexDefault(l *lexer) stateFn {
	pos := l.pos
	switch l.next() {
	case '\n':
		l.emit(item{
			typ:   lexNewLine,
			start: pos,
			end:   l.pos,
		})
		return lexDefault
	}
}

func lexText(l *lexer) stateFn {

}
