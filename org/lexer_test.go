package org

import (
	"testing"
)

type lexTest struct {
	name  string
	input string
	items []item
}

func mkItem(typ itemType, text string) item {
	return item{
		typ: typ,
		val: text,
	}
}

var (
	newLineItem = mkItem(itemNewline, "\n")
)

var testCases = []lexTest{
	{
		name:  "basic text",
		input: "This is some text",
		items: []item{
			mkItem(itemTextLine, "This is some text"),
		},
	},
	{
		name:  "leading space and text",
		input: "   This is some text",
		items: []item{
			mkItem(itemLeadingSpace, "   "),
			mkItem(itemTextLine, "This is some text"),
		},
	},
	{
		name:  "leading tabs and text",
		input: "\t\tThis is some text",
		items: []item{
			mkItem(itemLeadingSpace, "\t\t"),
			mkItem(itemTextLine, "This is some text"),
		},
	},
	{
		name:  "leading mixed spaces and text",
		input: "\t \t  This is some text",
		items: []item{
			mkItem(itemLeadingSpace, "\t \t  "),
			mkItem(itemTextLine, "This is some text"),
		},
	},
	{
		name:  "basic comment",
		input: "# This is a comment",
		items: []item{
			mkItem(itemComment, "# This is a comment"),
		},
	},
	{
		name:  "bogus comment",
		input: "#This is a comment",
		items: []item{
			mkItem(itemTextLine, "#This is a comment"),
		},
	},
	{
		name:  "comment with leading space",
		input: "\t \t  # This is a comment",
		items: []item{
			mkItem(itemLeadingSpace, "\t \t  "),
			mkItem(itemComment, "# This is a comment"),
		},
	},
	{
		name: "two lines of text",
		input: `This is some text -- line 1
This is some text -- line 2`,
		items: []item{
			mkItem(itemTextLine, "This is some text -- line 1"),
			newLineItem,
			mkItem(itemTextLine, "This is some text -- line 2"),
		},
	},
	{
		name: "two comments",
		input: `# This is a comment # line 1
# This is a comment # line 2`,
		items: []item{
			mkItem(itemComment, "# This is a comment # line 1"),
			newLineItem,
			mkItem(itemComment, "# This is a comment # line 2"),
		},
	},
	{
		name: "text then comment",
		input: `This is some text -- line 1
# This is a comment`,
		items: []item{
			mkItem(itemTextLine, "This is some text -- line 1"),
			newLineItem,
			mkItem(itemComment, "# This is a comment"),
		},
	},
	{
		name: "comment then text",
		input: `# This is a comment
This is some text -- line 2`,
		items: []item{
			mkItem(itemComment, "# This is a comment"),
			newLineItem,
			mkItem(itemTextLine, "This is some text -- line 2"),
		},
	},
	{
		name:  "basic heading",
		input: "* This is a heading",
		items: []item{
			mkItem(itemHeading, "* This is a heading"),
		},
	},
	{
		name: "not headings",
		input: `*This is NOT a heading
 * This is also NOT a heading`,
		items: []item{
			mkItem(itemTextLine, "*This is NOT a heading"),
			newLineItem,
			mkItem(itemLeadingSpace, " "),
			mkItem(itemTextLine, "* This is also NOT a heading"),
		},
	},
	{
		name: "bunch o headings",
		input: `* This is a heading 1
** This is a heading 2
*** This is a heading 3
**** This is a heading 4
***** This is a heading 5
****** This is a heading 6
******* This is a heading 7
`,
		items: []item{
			mkItem(itemHeading, "* This is a heading 1"),
			newLineItem,
			mkItem(itemHeading, "** This is a heading 2"),
			newLineItem,
			mkItem(itemHeading, "*** This is a heading 3"),
			newLineItem,
			mkItem(itemHeading, "**** This is a heading 4"),
			newLineItem,
			mkItem(itemHeading, "***** This is a heading 5"),
			newLineItem,
			mkItem(itemHeading, "****** This is a heading 6"),
			newLineItem,
			mkItem(itemHeading, "******* This is a heading 7"),
			newLineItem,
		},
	},
	{
		name:  "basic keyword",
		input: "#+KEY: VALUE",
		items: []item{
			mkItem(itemKeyword, "#+KEY: VALUE"),
		},
	},
}

func TestLexer(t *testing.T) {
	for _, v := range testCases {
		t.Run(v.name, func(t *testing.T) {
			l := lex(v.name, v.input)
			for _, expected := range v.items {
				r := <-l.items
				if r.typ != expected.typ {
					t.Errorf("Mismatched types: %v != %v", r.typ, expected.typ)
				}
				if r.val != expected.val {
					t.Errorf("Mismatched values: %s != %s", r.val, expected.val)
				}
			}
			r := <-l.items
			if r.typ != itemEOF {
				t.Errorf("Expected EOF, got: %v", r.typ)
			}
			if r.val != "" {
				t.Errorf("Expected empty string, got: %s", r.val)
			}
			r = <-l.items
			if r.typ != itemError {
				t.Errorf("Expected Error, got: %v", r.typ)
			}
		})
	}
}
