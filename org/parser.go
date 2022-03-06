package org

type nodeType int

const (
	nodeError nodeType = iota
	nodeComment
	nodeHeading
	nodePropertyDrawer
)

type parser struct {
	name  string
	lexer *lexer
}

func parse(name, input string) *parser {
	p := &parser{
		name:  name,
		lexer: lex(name, input),
	}
	return p
}
