package main

import (
	"reflect"
	"testing"
)

func TestSplitArgs(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
	}{
		{name: "empty", in: "", want: nil},
		{name: "whitespace only", in: "   \t  ", want: nil},
		{name: "single arg", in: "filesystem", want: []string{"filesystem"}},
		// documented space-separated form
		{name: "space separated", in: "-y @modelcontextprotocol/server-filesystem .",
			want: []string{"-y", "@modelcontextprotocol/server-filesystem", "."}},
		// quoting groups a token that contains spaces
		{name: "double quoted spaces", in: `a "b c" d`, want: []string{"a", "b c", "d"}},
		{name: "single quoted spaces", in: `'a b' c`, want: []string{"a b", "c"}},

		// Regression: a comma in a single token must NOT be treated as a
		// separator. Before the fix the comma heuristic split these into
		// bogus multiple args; now a comma is an ordinary character.
		{name: "comma token stays one arg", in: "a,b,c", want: []string{"a,b,c"}},
		{name: "url query with comma stays one arg", in: "http://x/api?a=1,b=2",
			want: []string{"http://x/api?a=1,b=2"}},
		{name: "comma tokens split only on space", in: "a,b c,d",
			want: []string{"a,b", "c,d"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := splitArgs(tc.in)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("splitArgs(%q) = %#v, want %#v", tc.in, got, tc.want)
			}
		})
	}
}
