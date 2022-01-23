package org

import "fmt"

type NodeType int

const (
	Keyword NodeType = iota
	Include
	Comment
	NodeWithMeta
	NodeWithName
	Headline
	Block
	Result
	InlineBlock
	Example
	Drawer
	List
	ListItem
	DescriptiveListItem
	Table
	HorizontalRule
	Paragraph
	Text
	Emphasis
	LatexFragment
	StatisticToken
	ExplicitLineBreak
	LineBreak
	RegularLink
	Macro
	Timestamp
	FootnoteLink
	FootnoteDefinition
)

type ParserError struct {
	message string
	line    int
	offset  int
}

// Error to conform to the Error interface
func (e *ParserError) Error() string {
	return fmt.Sprintf("parse error: %s (%d:%d)", e.message, e.line, e.offset)
}

type Point struct {
	Line   int
	Offset int
}

type Span struct {
	Start Point
	End   Point
	Body  string
}

type Node interface {
	Span() Span
	Type() NodeType
}

type Parser struct {
	out  chan Node
	lint int
	char int
}

// NewParser creates a new Parser
func NewParser() (*Parser, <-chan Node) {
	c := make(chan Node)
	return &Parser{
		out: c,
	}, c
}

// Reset resets the parser to the initial state
func (p *Parser) Reset() {

}

// Consume takes in a fragment of an Org doc and tries to parse it.
// If the fragment results in a production you can check on the output
// channel.
func (p *Parser) Consume(in string) error {

}
