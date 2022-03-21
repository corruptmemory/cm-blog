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

const (
	// Tag marking a substree as archived.
	archiveTag = "ARCHIVE"
	// Keyword matching a clock line.
	clockKeyword = "CLOCK:"
	// Keyword used to close TODO entries.
	closedKeyword = "CLOSED:"
	// Keyword used to mark deadline entries.
	deadlineKeyword = "DEADLINE:"
	// Keyword used to mark scheduled entries.
	scheduledKeyword  = "SCHEDULED:"
	dynamicBlockstart = "#+BLOCK:"
)

var (
	// Regexp matching any planning line keyword.
	planningKeywords = []string{
		scheduledKeyword,
		deadlineKeyword,
		closedKeyword,
	}
)

func planningLine(in string) (keyword string, value string, ok bool) {
	for _, v := range planningKeywords {
		ok = strings.HasPrefix(in, v+":")
		if ok {
			keyword = v
			value = strings.TrimSpace(in[strings.Index(in, ":")+1:])
			return
		}
	}
	return
}

func containsOnly(in string, chars string) bool {
	for _, c := range in {
		if !strings.ContainsRune(chars, c) {
			return false
		}
	}

	return true
}

func openingOrClosingDrawerLine(in string) (name string, ok bool) {
	p := strings.Split(in, ":")
	if len(p) == 3 && containsOnly(p[2], " \t") {
		return p[1], true
	}
	return
}

func dynamicBlockOpen(in string) (name, params string, err error) {
	if strings.HasPrefix(in, dynamicBlockstart) {
		p := strings.SplitN(strings.TrimSpace(strings.TrimPrefix(in, dynamicBlockstart)), " ", 2)
		switch len(p) {
		case 1:
			if p[0] == "" {
				return "", "", fmt.Errorf("start of dynamic block, but no name supplied")
			}
			return p[0], "", nil
		case 2:
			return p[0], p[1], nil
		}
	}
	return
}

func parseHeadline(in string) (level int, headline string, err error) {
	for i, c := range in {
		if c == '*' {
			level++
			continue
		}
		if c == ' ' {
			return level, strings.TrimSpace(in[i:]), nil
		}
		break
	}
	return 0, "", fmt.Errorf("could not parse headline from: %s", in)
}

func isalnum(c rune) bool {
	return '0' <= c && c <= '9' || 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z'
}

func parseFootnote(in string) (anchor string, ok bool) {
	if len(in) < 5 || !strings.HasPrefix(in, "[fn:") {
		return
	}
	tail := in[4:]
	for i, c := range tail {
		if isalnum(c) || c == '-' || c == '_' {
			continue
		}
		if c != ']' || i == 0 {
			return
		}
		return tail[:i], true
	}
	return
}

// (defvar org-element-paragraph-separate nil
//  "Regexp to separate paragraphs in an Org buffer.
// In the case of lines starting with \"#\" and \":\", this regexp
// is not sufficient to know if point is at a paragraph ending.  See
// `org-element-paragraph-parser' for more information.")
//
// (defvar org-element--object-regexp nil
//  "Regexp possibly matching the beginning of an object.
// This regexp allows false positives.  Dedicated parser (e.g.,
// `org-export-bold-parser') will take care of further filtering.
// Radio links are not matched by this regexp, as they are treated
// specially in `org-element--object-lex'.")
//
// (defun org-element--set-regexps ()
//  "Build variable syntax regexps."
//  (setq org-element-paragraph-separate
//	(concat "^\\(?:"
//		;; Headlines, inlinetasks.
//		"\\*+ " "\\|"
//		;; Footnote definitions.
//		"\\[fn:[-_[:word:]]+\\]" "\\|"
//		;; Diary sexps.
//		"%%(" "\\|"
//		"[ \t]*\\(?:"
//		;; Empty lines.
//		"$" "\\|"
//		;; Tables (any type).
//		"|" "\\|"
//		"\\+\\(?:-+\\+\\)+[ \t]*$" "\\|"
//		;; Comments, keyword-like or block-like constructs.
//		;; Blocks and keywords with dual values need to be
//		;; double-checked.
//		"#\\(?: \\|$\\|\\+\\(?:"
//		"BEGIN_\\S-+" "\\|"
//		"\\S-+\\(?:\\[.*\\]\\)?:[ \t]*\\)\\)"
//		"\\|"
//		;; Drawers (any type) and fixed-width areas.  Drawers
//		;; need to be double-checked.
//		":\\(?: \\|$\\|[-_[:word:]]+:[ \t]*$\\)" "\\|"
//		;; Horizontal rules.
//		"-\\{5,\\}[ \t]*$" "\\|"
//		;; LaTeX environments.
//		"\\\\begin{\\([A-Za-z0-9*]+\\)}" "\\|"
//		;; Clock lines.
//		"CLOCK:" "\\|"
//		;; Lists.
//		(let ((term (pcase org-plain-list-ordered-item-terminator
//			      (?\) ")") (?. "\\.") (_ "[.)]")))
//		      (alpha (and org-list-allow-alphabetical "\\|[A-Za-z]")))
//		  (concat "\\(?:[-+*]\\|\\(?:[0-9]+" alpha "\\)" term "\\)"
//			  "\\(?:[ \t]\\|$\\)"))
//		"\\)\\)")
//	org-element--object-regexp
//	(mapconcat #'identity
//		   (let ((link-types (regexp-opt (org-link-types))))
//		     (list
//		      ;; Sub/superscript.
//		      "\\(?:[_^][-{(*+.,[:alnum:]]\\)"
//		      ;; Bold, code, italic, strike-through, underline
//		      ;; and verbatim.
//                      (rx (or "*" "~" "=" "+" "_" "/") (not space))
//		      ;; Plain links.
//		      (concat "\\<" link-types ":")
//		      ;; Objects starting with "[": citations,
//		      ;; footnote reference, statistics cookie,
//		      ;; timestamp (inactive) and regular link.
//		      (format "\\[\\(?:%s\\)"
//			      (mapconcat
//			       #'identity
//			       (list "cite[:/]"
//				     "fn:"
//				     "\\(?:[0-9]\\|\\(?:%\\|/[0-9]*\\)\\]\\)"
//				     "\\[")
//			       "\\|"))
//		      ;; Objects starting with "@": export snippets.
//		      "@@"
//		      ;; Objects starting with "{": macro.
//		      "{{{"
//		      ;; Objects starting with "<" : timestamp
//		      ;; (active, diary), target, radio target and
//		      ;; angular links.
//		      (concat "<\\(?:%%\\|<\\|[0-9]\\|" link-types "\\)")
//		      ;; Objects starting with "$": latex fragment.
//		      "\\$"
//		      ;; Objects starting with "\": line break,
//		      ;; entity, latex fragment.
//		      "\\\\\\(?:[a-zA-Z[(]\\|\\\\[ \t]*$\\|_ +\\)"
//		      ;; Objects starting with raw text: inline Babel
//		      ;; source block, inline Babel call.
//		      "\\(?:call\\|src\\)_"))
//		   "\\|")))
//
// (org-element--set-regexps)
//
// ;;;###autoload
// (defun org-element-update-syntax ()
//  "Update parser internals."
//  (interactive)
//  (org-element--set-regexps)
//  (org-element-cache-reset 'all))
//
// (defconst org-element-all-elements
//  '(babel-call center-block clock comment comment-block diary-sexp drawer
//	       dynamic-block example-block export-block fixed-width
//	       footnote-definition headline horizontal-rule inlinetask item
//	       keyword latex-environment node-property paragraph plain-list
//	       planning property-drawer quote-block section
//	       special-block src-block table table-row verse-block)
//  "Complete list of element types.")
//
// (defconst org-element-greater-elements
//  '(center-block drawer dynamic-block footnote-definition headline inlinetask
//		 item plain-list property-drawer quote-block section
//		 special-block table org-data)
//  "List of recursive element types aka Greater Elements.")
//
// (defconst org-element-all-objects
//  '(bold citation citation-reference code entity export-snippet
//	 footnote-reference inline-babel-call inline-src-block italic line-break
//	 latex-fragment link macro radio-target statistics-cookie strike-through
//	 subscript superscript table-cell target timestamp underline verbatim)
//  "Complete list of object types.")
//
// (defconst org-element-recursive-objects
//  '(bold citation footnote-reference italic link subscript radio-target
//	 strike-through superscript table-cell underline)
//  "List of recursive object types.")
//
// (defconst org-element-object-containers
//  (append org-element-recursive-objects '(paragraph table-row verse-block))
//  "List of object or element types that can directly contain objects.")
//
// (defconst org-element-affiliated-keywords
//  '("CAPTION" "DATA" "HEADER" "HEADERS" "LABEL" "NAME" "PLOT" "RESNAME" "RESULT"
//    "RESULTS" "SOURCE" "SRCNAME" "TBLNAME")
//  "List of affiliated keywords as strings.
// By default, all keywords setting attributes (e.g., \"ATTR_LATEX\")
// are affiliated keywords and need not to be in this list.")
//
// (defconst org-element-keyword-translation-alist
//  '(("DATA" . "NAME")  ("LABEL" . "NAME") ("RESNAME" . "NAME")
//    ("SOURCE" . "NAME") ("SRCNAME" . "NAME") ("TBLNAME" . "NAME")
//    ("RESULT" . "RESULTS") ("HEADERS" . "HEADER"))
//  "Alist of usual translations for keywords.
// The key is the old name and the value the new one.  The property
// holding their value will be named after the translated name.")
//
// (defconst org-element-multiple-keywords '("CAPTION" "HEADER")
//  "List of affiliated keywords that can occur more than once in an element.

type node struct {
	nodeType nodeType
	parent   *node
	children []*node
	pos      int
	line     int
	level    int
	leading  string
	text     string
	keywords map[string]string
	drawer   map[string]string
}

func (n *node) addChild(child *node) {
	child.parent = n
	n.children = append(n.children, child)
}

func (n *node) addText(pos, line int, leading string, text string) {
	n.addChild(&node{
		nodeType: nodeText,
		pos:      pos,
		line:     line,
		leading:  leading,
		text:     text,
	})
}

func (n *node) addHeading(pos, line int, level int, heading string) {
	n.addChild(&node{
		nodeType: nodeHeading,
		pos:      pos,
		line:     line,
		level:    level,
		text:     heading,
	})
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
			}
			current.addChild(n)
			err = processItems(n, l)
			if err != nil {
				return
			}
		case itemKeyword:
			n := &node{
				nodeType: nodeKeyword,
			}
			current.addChild(n)
			err = processItems(n, l)
			if err != nil {
				return
			}
		case itemTextLine:
			n := &node{
				nodeType: nodeText,
			}
			current.addChild(n)
			err = processItems(n, l)
			if err != nil {
				return
			}
		case itemNodeMarker:
			n := &node{
				nodeType: nodeDrawer,
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

func withLeadingSpace(leading item, current *node, l *lexer) (err error) {
	i, ok := <-l.items
	if ok && i.typ != itemEOF {
		switch i.typ {
		case itemKeyword:
			current.addChild(&node{nodeType: nodeKeyword})
		case itemNewline, itemTextLine:
			current.addChild(&node{nodeType: nodeText})
		case itemNodeMarker:
			current.addChild(&node{nodeType: nodeDrawer})
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
