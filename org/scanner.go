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

type itemType int

const (
	textItem itemType = iota
	keywordItem
	headlineItem
	commentItem
)

type point struct {
	line   int
	offset int
}

func (p point) String() string {
	return fmt.Sprintf("(%d:%d)", p.line, p.offset)
}

type span struct {
	start point
	end   point
}

func (s span) String() string {
	return fmt.Sprintf("%s - %s", s.start, s.end)
}

type Item struct {
	itemType itemType
	span     span
	headline
	keyword
	comment
	text
}

func (i Item) String() string {
	switch i.itemType {
	case textItem:
		return fmt.Sprintf("Text[%s](%s)", i.span.String(), i.text.body)
	case headlineItem:
		return fmt.Sprintf("Headline[%s](%d, %s, %v)", i.span.String(), i.headline.level, i.headline.body, i.headline.tags)
	case commentItem:
		return fmt.Sprintf("Comment[%s](%s)", i.span.String(), i.comment.body)
	case keywordItem:
		return fmt.Sprintf("Keyword[%s](%s: %s)", i.span.String(), i.keyword.keyword, i.keyword.value)
	}
	return "<invalid>"
}

type headline struct {
	level int
	body  string
	tags  []string
}

type keyword struct {
	keyword string
	value   string
}

type comment struct {
	body string
}

type text struct {
	body string
}

type Scanner struct {
	input      string
	pos        int
	lineNumber int
	lineLen    int
	line       string
	eof        bool
	items      chan Item
}

func (s *Scanner) nextLine() bool {
	if s.eof {
		s.line = ""
		s.lineLen = 0
		return false
	}
	start := s.pos
	p := start
	for ; p < len(s.input) && s.input[p] != '\n'; p++ {
	}
	s.line = s.input[start:p]
	s.lineNumber++
	if p < len(s.input) {
		s.pos = p + 1
		s.lineLen = p - start
	} else {
		s.pos = p
		s.lineLen = p - start
		s.eof = true
	}
	return true
}

func (s *Scanner) next() {

}

func (s *Scanner) reset() {
	s.pos = 0
	s.line = ""
	s.lineNumber = 0
	s.lineLen = 0
	s.eof = false
}

func (s *Scanner) String() string {
	return fmt.Sprintf("[%d:%d:%d]: %s\n", s.lineNumber, s.pos, s.lineLen, s.line)
}

func (s *Scanner) withString(in string) {
	s.reset()
	s.input = in
}

func NewScanner(in string) (*Scanner, chan Item) {
	out := make(chan Item, 1000)
	r := &Scanner{
		input: in,
		items: out,
	}
	r.withString(in)
	return r, out
}

func isTagChar(r byte) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' || r == '@'
}

func (s *Scanner) currentPoint(offset int) point {
	return point{
		line:   s.lineNumber,
		offset: offset,
	}
}

func (s *Scanner) currentLine() int {
	return s.lineNumber
}

func (s *Scanner) findWhitespaceDelimitedString(start, end int) string {
	return strings.TrimSpace(s.line[start:end])
}

func (s *Scanner) currentWholeLine() span {
	return span{
		start: s.currentPoint(0),
		end:   s.currentPoint(s.lineLen),
	}
}

type scanFn func(scanner *Scanner) scanFn

func tryComment(s *Scanner) scanFn {
	m := commentRegexp.FindStringSubmatch(s.line)
	if len(m) == 0 {
		return nil
	}

	body := bytes.Buffer{}
	commentSpan := s.currentWholeLine()
	body.WriteString(m[1])
	var continueComment scanFn
	continueComment = func(s *Scanner) scanFn {
		if s.eof {
			return nil
		}

		m := commentRegexp.FindStringSubmatch(s.line)
		if len(m) == 0 {
			commentSpan.end.line = s.lineNumber
			commentSpan.end.offset = s.lineLen
			i := Item{
				itemType: commentItem,
				span:     commentSpan,
				comment: comment{
					body: body.String(),
				},
			}
			s.items <- i
			return initialState
		}

		body.WriteByte('\n')
		body.WriteString(m[1])
		return continueComment
	}
	return continueComment
}

func tryHeadline(s *Scanner) scanFn {
	stars := 0
	var pos int
	var r rune
	for pos, r = range s.line {
		if r == '*' {
			stars++
			continue
		}
		if stars == 0 || r != ' ' {
			return nil
		}
		break
	}
	if stars == 0 {
		return nil
	}
	headlineBody := ""
	var headlineTags []string
	in := s.findWhitespaceDelimitedString(pos, s.lineLen)
	if len(in) == 0 {
		return nil
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
	i := Item{
		itemType: headlineItem,
		span:     s.currentWholeLine(),
		headline: headline{
			level: stars,
			body:  headlineBody,
			tags:  headlineTags,
		},
	}
	s.items <- i
	return initialState
}

func tryKeyword(s *Scanner) scanFn {
	m := keywordRegexp.FindStringSubmatch(s.line)
	if len(m) == 0 {
		return nil
	}
	keywordKey := m[1]
	keywordValue := m[2]
	i := Item{
		itemType: keywordItem,
		span:     s.currentWholeLine(),
		keyword: keyword{
			keyword: keywordKey,
			value:   keywordValue,
		},
	}
	s.items <- i
	return initialState
}

func otherThanText(s *Scanner) scanFn {
	var next scanFn
	next = tryComment(s)
	if next != nil {
		return next
	}
	next = tryHeadline(s)
	if next != nil {
		return next
	}
	next = tryKeyword(s)
	if next != nil {
		return next
	}
	return nil
}

func tryText(s *Scanner) scanFn {
	body := bytes.Buffer{}
	textSpan := s.currentWholeLine()
	body.WriteString(s.line)
	var continueText scanFn
	continueText = func(s *Scanner) scanFn {
		emit := func() {
			textSpan.end.line = s.currentLine()
			textSpan.end.offset = s.lineLen
			i := Item{
				itemType: textItem,
				span:     textSpan,
				text: text{
					body: body.String(),
				},
			}
			s.items <- i
		}

		ott := otherThanText(s)
		if ott != nil {
			emit()
			return ott
		}
		if s.eof {
			emit()
			return nil
		}
		body.WriteByte('\n')
		body.WriteString(s.line)
		return continueText
	}
	return continueText
}

func initialState(s *Scanner) scanFn {
	if s.eof {
		return nil
	} else if s.pos >= len(s.input) {
		s.eof = true
		s.pos = len(s.input)
		s.line = ""
		return nil
	}

	var next scanFn
	next = tryComment(s)
	if next != nil {
		return next
	}
	next = tryHeadline(s)
	if next != nil {
		return next
	}
	next = tryKeyword(s)
	if next != nil {
		return next
	}
	return tryText(s)
}

func (s *Scanner) Scan() error {
	s.nextLine()
	for next := initialState(s); next != nil; next = next(s) {
		s.nextLine()
	}
	close(s.items)
	return nil
}
