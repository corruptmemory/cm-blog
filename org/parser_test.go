package org

import (
	"fmt"
	"testing"
)

type parserTest struct {
	name  string
	input string
}

var testParserCases = []parserTest{
	{
		name:  "basic text",
		input: "This is some text",
	},
	{
		name:  "leading space and text",
		input: "   This is some text",
	},
	{
		name:  "leading tabs and text",
		input: "\t\tThis is some text",
	},
	{
		name:  "leading mixed spaces and text",
		input: "\t \t  This is some text",
	},
	{
		name:  "basic comment",
		input: "# This is a comment",
	},
	{
		name:  "bogus comment",
		input: "#This is a comment",
	},
	{
		name:  "comment with leading space",
		input: "\t \t  # This is a comment",
	},
	{
		name: "two lines of text",
		input: `This is some text -- line 1
This is some text -- line 2`,
	},
	{
		name: "two comments",
		input: `# This is a comment # line 1
# This is a comment # line 2`,
	},
	{
		name: "text then comment",
		input: `This is some text -- line 1
# This is a comment`,
	},
	{
		name: "comment then text",
		input: `# This is a comment
This is some text -- line 2`,
	},
	{
		name:  "basic heading",
		input: "* This is a heading",
	},
	{
		name: "not headings",
		input: `*This is NOT a heading
 * This is also NOT a heading`,
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
	},
	{
		name:  "basic keyword",
		input: "#+KEY: VALUE",
	},
	{
		name:  "leading space then keyword",
		input: "   #+KEY: VALUE",
	},
	{
		name:  "leading space then keyword",
		input: "   #+KEY: VALUE",
	},
	{
		name:  "not a keyword",
		input: "#+ KEY: VALUE",
	},
	{
		name:  "basic drawer",
		input: ":DRAWER:",
	},
	{
		name:  "space then drawer",
		input: "   :DRAWER:",
	},
	{
		name:  "setting node",
		input: ":SETTING: Value",
	},
	{
		name:  "setting node with plus",
		input: ":SETTING+: Value",
	},
	{
		name:  "flag",
		input: ":FLAG+:",
	},
	{
		name: "drawer",
		input: `:MY_BLOCK:
This is some text
:A-THINGY: With a value   
:REALLY-TEXT:As you can see
# But I can comment anywhere
:END:
`,
	},
}

func TestParser(t *testing.T) {
	for _, v := range testParserCases {
		t.Run(v.name, func(t *testing.T) {
			n, err := parse(v.name, v.input)
			if err != nil {
				t.Error(err)
				t.Fail()
			}
			fmt.Printf("n: %v", n)
		})
	}
}
