package org

import (
	"fmt"
	"strings"
)

type nodeType int

const (
	nodeError nodeType = iota
	nodeDocument
	nodeText
	nodeKeyword
	nodeHeading
	nodeDrawer
)

type node struct {
	nodeType nodeType
	parent   *node
	children []*node
}

func (n *node) addChild(child *node) {
	n.children = append(n.children, child)
}

func (n *node) String() string {
	r := ""
	switch n.nodeType {
	case nodeDocument:
		r = "nodeDocument"
	case nodeText:
		r = "nodeText"
	case nodeKeyword:
		r = "nodeKeyword"
	case nodeHeading:
		r = "nodeHeading"
	case nodeDrawer:
		r = "nodeDrawer"
	}
	var c []string
	for _, i := range n.children {
		c = append(c, i.String())
	}
	return fmt.Sprintf("%s: %p { %s }", r, n.parent, strings.Join(c, ", "))
}

func processItems(current *node, l *lexer) (err error) {
	for i := range l.items {
		switch i.typ {
		case itemEOF, itemError:
			return nil
		case itemHeading:
			n := &node{
				nodeType: nodeHeading,
				parent:   current,
			}
			current.addChild(n)
			err = processItems(n, l)
			if err != nil {
				return
			}
		case itemKeyword:
			n := &node{
				nodeType: nodeKeyword,
				parent:   current,
			}
			current.addChild(n)
			err = processItems(n, l)
			if err != nil {
				return
			}
		case itemTextLine:
			n := &node{
				nodeType: nodeText,
				parent:   current,
			}
			current.addChild(n)
			err = processItems(n, l)
			if err != nil {
				return
			}
		case itemNodeMarker:
			n := &node{
				nodeType: nodeDrawer,
				parent:   current,
			}
			current.addChild(n)
			err = processItems(n, l)
			if err != nil {
				return
			}
		default:
			// skip
		}
	}
	return
}

func parse(name, input string) (*node, error) {
	l := lex(name, input)
	doc := node{
		nodeType: nodeDocument,
	}
	err := processItems(&doc, l)
	if err != nil {
		return nil, err
	}
	return &doc, nil
}
