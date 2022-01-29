package org

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

var (
	keywordRegexp = regexp.MustCompile("^[ \t]*#\\+([0-9a-zA-Z_@\\[\\]]+):([^\n]*)")
	commentRegexp = regexp.MustCompile("^[ \t]*#([^+][^\n]*)")
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

type Scanner struct {
	buffer  []byte
	Pos     int
	Line    int
	LineLen int
	LineBuf []byte
	EOF     bool
}

func (s *Scanner) NextLine() bool {
	if s.EOF {
		s.LineBuf = nil
		s.LineLen = 0
		return false
	}
	start := s.Pos
	p := start
	for ; p < len(s.buffer) && s.buffer[p] != '\n'; p++ {
	}
	s.LineBuf = s.buffer[start:p]
	s.Line++
	if p < len(s.buffer) {
		s.Pos = p + 1
		s.LineLen = p - start
	} else {
		s.Pos = p
		s.LineLen = p - start
		s.EOF = true
	}
	return true
}

func (s *Scanner) Reset() {
	s.buffer = s.buffer[:0]
	s.Pos = 0
	s.Line = 0
	s.LineLen = 0
	s.LineBuf = nil
	s.EOF = false
}

func (s *Scanner) String() string {
	return fmt.Sprintf("[%d:%d:%d]: %s\n", s.Line, s.Pos, s.LineLen, string(s.LineBuf))
}

func (s *Scanner) WithBytes(in []byte) {
	s.Reset()
	s.buffer = append(s.buffer, in...)
}

func NewScanner(in []byte) *Scanner {
	r := &Scanner{}
	r.WithBytes(in)
	return r
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
	String() string
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

type TextNode struct {
	span Span
	buf  *bytes.Buffer
	Body string
}

func (t *TextNode) Span() Span {
	return t.span
}

func (t *TextNode) String() string {
	return fmt.Sprintf("Text[%s; %s]", t.Body, t.span)
}

type Parser struct {
	out     chan Node
	scanner Scanner
	state   State
}

func NewParser() (*Parser, <-chan Node) {
	c := make(chan Node, 100)
	return &Parser{
		out: c,
	}, c
}

// Reset resets the parser to the initial state
func (p *Parser) Reset() {
	p.scanner.Reset()
	p.state = empty
}

func isTagChar(r byte) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '@'
}

func (p *Parser) currentPoint(offset int) Point {
	return Point{
		Line:   p.scanner.Line,
		Offset: offset,
	}
}

func (p *Parser) currentLine() int {
	return p.scanner.Line
}

func isHorizontalWhitespace(in byte) bool {
	switch in {
	case ' ', '\t':
		return true
	default:
		return false
	}
}

func (p *Parser) findWhitespaceDelimitedString(start, end int) string {
	s := start
	for ; isHorizontalWhitespace(p.scanner.LineBuf[s]); s++ {
	}
	e := end
	for ; isHorizontalWhitespace(p.scanner.LineBuf[e]); e-- {
	}
	return string(p.scanner.LineBuf[s:e])
}

func (p *Parser) currentWholeLine() Span {
	return Span{
		Start: p.currentPoint(0),
		End:   p.currentPoint(p.scanner.LineLen),
	}
}

type lexer struct {
}

type scanFn func(*lexer) scanFn

func (p *Parser) Parse(file string, in []byte) error {
	if len(in) == 0 {
		return nil
	}
	p.scanner.WithBytes(in)

	var node Node = nil
	state := empty

	closeMultiLineStructure := func() {}

	var keywordKey string
	var keywordValue string

	isKeywordLine := func() bool {
		m := keywordRegexp.FindSubmatch(p.scanner.LineBuf)
		if len(m) > 0 {
			keywordKey = string(m[1])
			keywordValue = string(m[2])
			return true
		}
		return false
	}

	var commentBody string

	isCommentLine := func() bool {
		m := commentRegexp.FindSubmatch(p.scanner.LineBuf)
		if len(m) > 0 {
			commentBody = string(m[1])
			return true
		}
		return false
	}

	var headlineLevel int
	var headlineBody string
	var headlineTags []string

	isHeadlineLine := func() bool {
		stars := 0
		var pos int
		var r rune
		for pos, r = range p.scanner.LineBuf {
			if r == '*' {
				stars++
				continue
			}
			if stars == 0 || r != ' ' {
				return false
			}
		}
		headlineLevel = stars
		headlineBody = ""
		headlineTags = nil
		in := p.findWhitespaceDelimitedString(pos, p.scanner.LineLen)
		if len(in) == 0 {
			return false
		}
		if in[len(in)-1] == ':' {
			lastColon := len(in) - 1
			for p := lastColon - 1; p > 0; p-- {
				c := in[p]
				if isTagChar(c) {
					continue
				} else if c == ':' {
					if lastColon-p > 1 {
						headlineTags = append(headlineTags, in[p+1:lastColon])
					}
					lastColon = p
				} else {
					break
				}
			}
			headlineBody = strings.TrimSpace(in[:lastColon])
		} else {
			headlineBody = in
		}
		return true
	}

	for p.scanner.NextLine() {
		switch {
		case isKeywordLine():
			closeMultiLineStructure()
			state = empty
			p.out <- &KeywordNode{
				span:    p.currentWholeLine(),
				Keyword: keywordKey,
				Value:   strings.TrimSpace(keywordValue),
			}
		case isCommentLine():
			var cn *CommentNode
			if state == comment {
				cn = node.(*CommentNode)
				cn.buf.WriteByte('\n')
				cn.span.End = p.currentPoint(0)
			} else {
				closeMultiLineStructure()
				state = comment
				cn = &CommentNode{
					span: p.currentWholeLine(),
					buf:  &bytes.Buffer{},
				}
				node = cn
			}
			cn.buf.WriteString(commentBody)
		case isHeadlineLine():
			closeMultiLineStructure()
			state = empty
			p.out <- &HeadlineNode{
				span:  p.currentWholeLine(),
				Level: headlineLevel,
				Body:  headlineBody,
				Tags:  headlineTags,
			}
		default:
			var tn *TextNode
			if state == text {
				tn = node.(*TextNode)
				tn.buf.WriteByte('\n')
				tn.span.End = p.currentPoint(0)
			} else {
				closeMultiLineStructure()
				state = text
				tn = &TextNode{
					span: p.currentWholeLine(),
					buf:  &bytes.Buffer{},
				}
				node = tn
			}
			tn.buf.Write(p.scanner.LineBuf)
		}
	}
	return nil
}
