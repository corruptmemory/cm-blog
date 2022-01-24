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
	c := make(chan Node)
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
		return i+1
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

func (p *Parser) nextLine()  {
  p.line++
}

func (p *Parser) resetBufferWith(in string)  {
  p.buffer = p.buffer[:0]
	p.appendInput(in)
}

func (p *Parser) appendInput(in string)  {
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
}

func (p *Parser) determineState() {
	if p.isHeadlinePrefix() {
		p.state = headline
		p.node = &HeadlineNode{
			span:  Span{
				Start: p.currentPoint(),
				End: Point{Line:p.currentLine()},
			},
		}
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
	for ;isSpace(
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
	switch p.state {
	case headline:
		eol := p.findEOL(0)
		if eol > 0 {
			node := p.node.(*HeadlineNode)
			node.span.End.Offset = eol-1
			node.Level = p.snagHeadline()
			
		}
	case text:
		
	}
	if in[0] == '*' {
		p.state = headline
		p.resetBufferWith(in)
		eol := p.findEOL(node.Level)
		if eol >= 0 {
		}
	}

	return nil
}

func (p *Parser) EOF() {
  
}
