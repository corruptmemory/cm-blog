package org

import (
	"bytes"
	"fmt"
)

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
}

type Node interface {
	Span() Span
	Type() NodeType
}

type State int

const (
	empty State = iota
	text
	headline
)

type HeadlineNode struct {
	span  Span
	Level int
	Body  string
}

func (h *HeadlineNode) Type() NodeType {
	return Headline
}

func (h *HeadlineNode) Span() Span {
	return h.span
}

func (h *HeadlineNode) String() string {
	return fmt.Sprintf("Headline[%d, '%s']", h.Level, h.Body)
}

type TextNode struct {
	span Span
	buf  *bytes.Buffer
	Body string
}

func (t *TextNode) Type() NodeType {
	return Text
}

func (t *TextNode) Span() Span {
	return t.span
}

func (t *TextNode) String() string {
	return fmt.Sprintf("Text[%s]", t.Body)
}

type Parser struct {
	out    chan Node
	line   int
	offset int
	state  State
	buffer []rune
	node   Node
}

// NewParser creates a new Parser
func NewParser() (*Parser, <-chan Node) {
	c := make(chan Node, 100)
	return &Parser{
		out: c,
	}, c
}

// Reset resets the parser to the initial state
func (p *Parser) Reset() {
	p.line = 0
	p.offset = 0
	p.state = empty
	p.buffer = p.buffer[:0]

}

func (p *Parser) snagHeadline() int {
	for i, c := range p.buffer {
		if c == '*' {
			continue
		}
		return i
	}
	return 0
}

func (p *Parser) currentPoint() Point {
	return Point{
		Line:   p.line + 1,
		Offset: p.offset + 1,
	}
}

func (p *Parser) currentLine() int {
	return p.line + 1
}

func (p *Parser) nextLine() {
	p.line++
}

func (p *Parser) appendInput(in string) {
	for _, r := range in {
		p.buffer = append(p.buffer, r)
	}
}

func (p *Parser) findEOL(from int) int {
	for i, r := range p.buffer[from:] {
		if r == '\n' {
			return i
		}
	}
	return -1
}

// isHeadlinePrefix ...
func (p *Parser) isHeadlinePrefix() bool {
	for _, r := range p.buffer {
		if r == '*' {
			continue
		}
		if r == ' ' {
			return true
		}
		return false
	}
	return false
}

func (p *Parser) determineState() {
	if len(p.buffer) == 0 {
		p.state = empty
		p.node = nil
		return
	}
	if p.isHeadlinePrefix() {
		p.state = headline
		p.node = &HeadlineNode{
			span: Span{
				Start: p.currentPoint(),
				End:   Point{Line: p.currentLine()},
			},
		}
		return
	}
	p.state = text
}

func isSpace(in rune) bool {
	switch in {
	case ' ', '\t':
		return true
	default:
		return false
	}
}

func (p *Parser) findWhitespaceDelimitedString(start, end int) string {
	s := start
	for ; isSpace(p.buffer[s]); s++ {
	}
	e := end
	for ; isSpace(p.buffer[e]); e-- {
	}
	return string(p.buffer[s:e])
}

func (p *Parser) advanceLine() {
	eol := p.findEOL(0)
	if eol >= 0 {
		p.nextLine()
		p.offset = 0
		eol++
		if eol < len(p.buffer) {
			tail := len(p.buffer[eol:])
			copy(p.buffer[0:], p.buffer[eol:])
			p.buffer = p.buffer[:tail]
		} else {
			p.buffer = p.buffer[:0]
		}
	}
}

func (p *Parser) consumeAll() {
	p.buffer = p.buffer[:0]
}

// Consume takes in a fragment of an Org doc and tries to parse it.
// If the fragment results in a production you can check on the output
// channel.
func (p *Parser) Consume(in string) error {
	if len(in) == 0 {
		return nil
	}
	p.appendInput(in)
	if p.state == empty {
		p.determineState()
	}
	for {
		if len(p.buffer) == 0 {
			return nil
		}
		switch p.state {
		case empty:
			return nil
		case headline:
			eol := p.findEOL(0)
			node := p.node.(*HeadlineNode)
			if eol > 0 {
				node.span.End.Offset = eol - 1
				node.Level = p.snagHeadline()
				node.Body = p.findWhitespaceDelimitedString(node.Level, eol)
				p.out <- node
				p.node = nil
				p.advanceLine()
				p.determineState()
				continue
			}
			return nil
		case text:
			if p.node == nil {
				p.node = &TextNode{
					span: Span{
						Start: p.currentPoint(),
					},
					buf: &bytes.Buffer{},
				}
			}
			eol := p.findEOL(0)
			node := p.node.(*TextNode)
			if eol >= 0 {
				node.span.End.Line = p.currentLine()
				node.span.End.Offset = eol
				node.buf.WriteString(string(p.buffer[0 : eol+1]))
				p.advanceLine()
				p.determineState()
				if p.state != text {
					node.Body = node.buf.String()
					node.buf = nil
					p.out <- node
				}
				continue
			} else {
				node.span.End.Line = p.currentLine()
				node.span.End.Offset = len(p.buffer)
				node.buf.WriteString(string(p.buffer[0:len(p.buffer)]))
				p.consumeAll()
			}
		}
	}
}

func (p *Parser) EOF() {
	switch p.state {
	case headline:
		node := p.node.(*HeadlineNode)
		node.span.End.Offset = len(p.buffer)
		node.Level = p.snagHeadline()
		node.Body = p.findWhitespaceDelimitedString(node.Level, len(p.buffer)-1)
		p.out <- node
	case text:
		node := p.node.(*TextNode)
		node.span.End.Line = p.currentLine()
		node.span.End.Offset = len(p.buffer)
		node.buf.WriteString(string(p.buffer[0:len(p.buffer)]))
		node.Body = node.buf.String()
		node.buf = nil
		p.out <- node
	}
	close(p.out)
}
