package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
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

// TestVersionLockstep is the single-source-of-truth guard for every user-visible
// version surface. The `version` package var is what `mcpx --version` reports
// (cobra's Version field). The VERSION file, the CHANGELOG head (`## [X.Y.Z]`),
// and web/site.json's `meta.content_version` must all agree with it. This test
// FAILS on the shipped v0.4.0 tag (the CHANGELOG head read [0.1.0] while the
// version var was 0.4.0, and web/site.json's meta.content_version was v0.4.0)
// and PASSES once the surfaces are put back in lockstep — proving the drift was
// real. The site field carries a leading "v" (e.g. "v0.5.0"); it is stripped
// before comparison so the bare version var is the single canonical form.
func TestVersionLockstep(t *testing.T) {
	// Locate the repo root from this test file's path (cmd/mcpx/main_test.go
	// sits two levels below the repo root).
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to locate the test source")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")

	// 1) VERSION file.
	versionBytes, err := os.ReadFile(filepath.Join(repoRoot, "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	fileVersion := strings.TrimSpace(string(versionBytes))

	// 2) CHANGELOG head: the first "## [X.Y.Z]" heading.
	changelogBytes, err := os.ReadFile(filepath.Join(repoRoot, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}
	var changelogHead string
	for _, line := range strings.Split(string(changelogBytes), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## [") {
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start >= 0 && end > start {
				changelogHead = line[start+1 : end]
				break
			}
		}
	}
	if changelogHead == "" {
		t.Fatal("no '## [version]' heading found in CHANGELOG.md")
	}

	// 3) web/site.json meta.content_version (carries a leading "v").
	siteBytes, err := os.ReadFile(filepath.Join(repoRoot, "web", "site.json"))
	if err != nil {
		t.Fatalf("read web/site.json: %v", err)
	}
	var site struct {
		Meta struct {
			ContentVersion string `json:"content_version"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(siteBytes, &site); err != nil {
		t.Fatalf("parse web/site.json: %v", err)
	}
	siteVersion := strings.TrimPrefix(site.Meta.ContentVersion, "v")

	want := version
	if fileVersion != want {
		t.Errorf("VERSION file = %q, want %q (version var)", fileVersion, want)
	}
	if changelogHead != want {
		t.Errorf("CHANGELOG head = %q, want %q (version var)", changelogHead, want)
	}
	if siteVersion != want {
		t.Errorf("web/site.json meta.content_version = %q, want %q (version var)", site.Meta.ContentVersion, want)
	}
}
