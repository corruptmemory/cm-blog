package org

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
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

var (
	keywordRegexp = regexp.MustCompile("^[ \t]*#\\+([0-9a-zA-Z_@\\[\\]]+):([^\n]*)")
	commentRegexp = regexp.MustCompile("^[ \t]*#([^+][^\n]*)")
)

type ScannerError struct {
	message string
	line    int
	offset  int
}

// Error to conform to the Error interface
func (e *ScannerError) Error() string {
	return fmt.Sprintf("parse error: %s (%d:%d)", e.message, e.line, e.offset)
}

type Point struct {
	Line   int
	Offset int
}

func (p Point) String() string {
	return fmt.Sprintf("(%d:%d)", p.Line, p.Offset)
}

type Span struct {
	Start Point
	End   Point
}

func (s Span) String() string {
	return fmt.Sprintf("%s - %s", s.Start, s.End)
}

type Node interface {
	Span() Span
	Type() NodeType
}

type State int

const (
	empty State = iota
	keyword
	text
	headline
	comment
)

type HeadlineNode struct {
	span  Span
	Level int
	Body  string
	Tags  []string
}

func (h *HeadlineNode) Type() NodeType {
	return Headline
}

func (h *HeadlineNode) Span() Span {
	return h.span
}

func (h *HeadlineNode) String() string {
	return fmt.Sprintf("Headline[%d, '%s', [%s]; %s]", h.Level, h.Body, strings.Join(h.Tags, ", "), h.span)
}

type KeywordNode struct {
	span    Span
	Keyword string
	Value   string
}

func (k *KeywordNode) String() string {
	return fmt.Sprintf("Keyword[Key: %s, Value: %s; %s]", k.Keyword, k.Value, k.span)
}

func (k *KeywordNode) Span() Span {
	return k.span
}

func (k *KeywordNode) Type() NodeType {
	return Keyword
}

type CommentNode struct {
	span Span
	buf  *bytes.Buffer
	Body string
}

func (c *CommentNode) String() string {
	return fmt.Sprintf("Comment[%s; %s]", c.Body, c.span)
}

func (c *CommentNode) Span() Span {
	return c.span
}

func (c *CommentNode) Type() NodeType {
	return Comment
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
	return fmt.Sprintf("Text[%s; %s]", t.Body, t.span)
}

type Scanner struct {
	out    chan Node
	line   int
	offset int
	state  State
	buffer []byte
	node   Node
}

// NewScanner creates a new Scanner
func NewScanner() (*Scanner, <-chan Node) {
	c := make(chan Node, 100)
	return &Scanner{
		out: c,
	}, c
}

// Reset resets the parser to the initial state
func (p *Scanner) Reset() {
	p.line = 0
	p.offset = 0
	p.state = empty
	p.buffer = p.buffer[:0]

}

func isTagChar(r byte) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '@'
}

func scanHeadline(in string) (headline string, tags []string) {
	if len(in) == 0 {
		return
	}
	if in[len(in)-1] != ':' {
		return in, nil
	}
	lastColon := len(in) - 1
	for p := lastColon - 1; p > 0; p-- {
		c := in[p]
		if isTagChar(c) {
			continue
		} else if c == ':' {
			if lastColon-p > 1 {
				tags = append(tags, in[p+1:lastColon])
			}
			lastColon = p
		} else {
			break
		}
	}
	headline = in[:lastColon]
	return
}

func (p *Scanner) snagHeadline() int {
	for i, c := range p.buffer {
		if c == '*' {
			continue
		}
		return i
	}
	return 0
}

func (p *Scanner) currentPoint() Point {
	return Point{
		Line:   p.line + 1,
		Offset: p.offset + 1,
	}
}

func (p *Scanner) currentLine() int {
	return p.line + 1
}

func (p *Scanner) nextLine() {
	p.line++
}

func (p *Scanner) appendInput(in string) {
	p.buffer = append(p.buffer, []byte(in)...)
}

func (p *Scanner) findEOL() int {
	for i, r := range p.buffer {
		if r == '\n' {
			return i
		}
	}
	return -1
}

func (p *Scanner) isHeadlinePrefix() bool {
	stars := 0
	for _, r := range p.buffer {
		if r == '*' {
			stars++
			continue
		}
		if stars > 0 && r == ' ' {
			return true
		}
		return false
	}
	return false
}

func (p *Scanner) isKeywordLine() (keyword string, value string, match bool) {
	m := keywordRegexp.FindSubmatch(p.buffer)
	if len(m) > 0 {
		return string(m[1]), string(m[2]), true
	}
	return
}

func (p *Scanner) isCommentLine() (body string, match bool) {
	m := commentRegexp.FindSubmatch(p.buffer)
	if len(m) > 0 {
		return string(m[1]), true
	}
	return
}

func (p *Scanner) containsLine() bool {
	return bytes.ContainsRune(p.buffer, '\n')
}

func (p *Scanner) determineState() {
	possiblyCloseText := func() {
		if p.state == text {
			node := p.node.(*TextNode)
			node.Body = node.buf.String()
			node.buf = nil
			p.state = empty
			p.out <- node
		}
	}
	if len(p.buffer) == 0 {
		p.state = empty
		p.node = nil
		return
	}
	if kw, v, kwl := p.isKeywordLine(); kwl {
		possiblyCloseText()
		p.state = keyword
		p.node = &KeywordNode{
			span: Span{
				Start: p.currentPoint(),
				End:   Point{Line: p.currentLine()},
			},
			Keyword: kw,
			Value:   strings.TrimSpace(v),
		}
		return
	}
	if cb, cl := p.isCommentLine(); cl {
		possiblyCloseText()
		p.state = comment
		buf := &bytes.Buffer{}
		buf.WriteString(strings.TrimSpace(cb))
		p.node = &CommentNode{
			span: Span{
				Start: p.currentPoint(),
				End:   Point{Line: p.currentLine()},
			},
			buf: buf,
		}
		return
	}
	if p.isHeadlinePrefix() {
		possiblyCloseText()
		p.state = headline
		p.node = &HeadlineNode{
			span: Span{
				Start: p.currentPoint(),
				End:   Point{Line: p.currentLine()},
			},
		}
		return
	}
	if p.state == text {
		return
	}
	p.state = text
	p.node = &TextNode{
		span: Span{
			Start: p.currentPoint(),
		},
		buf: &bytes.Buffer{},
	}
}

func isHorizontalWhitespace(in byte) bool {
	switch in {
	case ' ', '\t':
		return true
	default:
		return false
	}
}

func (p *Scanner) findWhitespaceDelimitedString(start, end int) string {
	s := start
	for ; isHorizontalWhitespace(p.buffer[s]); s++ {
	}
	e := end
	for ; isHorizontalWhitespace(p.buffer[e]); e-- {
	}
	return string(p.buffer[s:e])
}

func (p *Scanner) advanceLine() {
	eol := p.findEOL()
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

func (p *Scanner) consumeAll() {
	p.buffer = p.buffer[:0]
}

// Consume takes in a fragment of an Org doc and tries to parse it.
// If the fragment results in a production you can check on the output
// channel.
func (p *Scanner) Consume(in string) error {
	if len(in) == 0 {
		return nil
	}
	p.appendInput(in)
	for {
		if !p.containsLine() {
			return nil
		}
		eol := p.findEOL()
		p.determineState()
		switch p.state {
		case empty:
			return nil
		case keyword:
			node := p.node.(*KeywordNode)
			node.span.End.Offset = eol + 1
			p.out <- node
			p.node = nil
			p.advanceLine()
			p.state = empty
		case comment:
			node := p.node.(*CommentNode)
			node.span.End.Offset = eol + 1
			for {
				p.advanceLine()
				if p.containsLine() {
					eol = p.findEOL()
					if cb, cl := p.isCommentLine(); cl {
						node.buf.WriteByte('\n')
						node.buf.WriteString(cb)
						node.span.End.Line = p.currentLine()
						node.span.End.Offset = eol + 1
						continue
					}
					p.state = empty
					node.Body = node.buf.String()
					node.buf = nil
					p.out <- node
				}
				break
			}
		case headline:
			node := p.node.(*HeadlineNode)
			node.span.End.Offset = eol + 1
			node.Level = p.snagHeadline()
			hl, tags := scanHeadline(p.findWhitespaceDelimitedString(node.Level, eol))
			node.Body = hl
			node.Tags = tags
			p.out <- node
			p.node = nil
			p.advanceLine()
			p.state = empty
		case text:
			node := p.node.(*TextNode)
			for {
				node.span.End.Line = p.currentLine()
				node.span.End.Offset = eol + 1
				node.buf.WriteString(string(p.buffer[0 : eol+1]))
				p.advanceLine()
				if p.containsLine() {
					eol = p.findEOL()
					p.determineState()
					if p.state != text {
						break
					}
				} else {
					break
				}
			}
		}
	}
}

func (p *Scanner) EOF() {
	if p.state == empty {
		p.determineState()
	}
	eol := len(p.buffer)
	switch p.state {
	case empty:
	case keyword:
		node := p.node.(*KeywordNode)
		node.span.End.Offset = eol + 1
		p.out <- node
	case comment:
		node := p.node.(*CommentNode)
		node.span.End.Offset = eol + 1
		p.out <- node
	case headline:
		node := p.node.(*HeadlineNode)
		node.span.End.Offset = eol + 1
		node.Level = p.snagHeadline()
		hl, tags := scanHeadline(p.findWhitespaceDelimitedString(node.Level, eol-1))
		node.Body = hl
		node.Tags = tags
		p.out <- node
	case text:
		node := p.node.(*TextNode)
		node.span.End.Line = p.currentLine()
		node.span.End.Offset = eol + 1
		node.buf.WriteString(string(p.buffer[0:eol]))
		node.Body = node.buf.String()
		node.buf = nil
		p.out <- node
	}
	p.node = nil
	close(p.out)
}
