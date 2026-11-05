package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"runtime"
	"strings"
	"testing"

	mcpxver "github.com/SuperMarioYL/mcpx/internal/version"
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
// version surface. internal/version.Version is what `mcpx --version` reports
// (the ldflags-overridable `version` var defaults to it). The VERSION file, the
// CHANGELOG head (`## [X.Y.Z]`), web/site.json's `meta.content_version` and
// `meta.source_version`, and a release link reference for every CHANGELOG
// version heading must all agree with it. This test FAILS on the shipped v0.4.0
// tag (the CHANGELOG head read [0.1.0] while the version var was 0.4.0, and
// web/site.json's meta.content_version was v0.4.0) and PASSES once the surfaces
// are put back in lockstep — proving the drift was real. The site fields carry a
// leading "v" (e.g. "v0.5.0"); it is stripped before comparison so the bare
// version var is the single canonical form.
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

	// 2) CHANGELOG: the head version plus every version heading's release link.
	changelogBytes, err := os.ReadFile(filepath.Join(repoRoot, "CHANGELOG.md"))
	if err != nil {
		t.Fatalf("read CHANGELOG.md: %v", err)
	}
	var changelogHead string
	var headings []string
	for _, line := range strings.Split(string(changelogBytes), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "## [") {
			start := strings.Index(line, "[")
			end := strings.Index(line, "]")
			if start >= 0 && end > start {
				v := line[start+1 : end]
				headings = append(headings, v)
				if changelogHead == "" {
					changelogHead = v
				}
			}
		}
	}
	if changelogHead == "" {
		t.Fatal("no '## [version]' heading found in CHANGELOG.md")
	}
	// A heading without a link reference renders as dead text: the v0.5.0 entry
	// was added without its own [X.Y.0]: line, so the newest release was the only
	// unlinked one.
	referenced := map[string]bool{}
	for _, line := range strings.Split(string(changelogBytes), "\n") {
		if m := changelogLinkRef.FindStringSubmatch(strings.TrimSpace(line)); m != nil {
			referenced[m[1]] = true
		}
	}
	for _, v := range headings {
		if !referenced[v] {
			t.Errorf("CHANGELOG heading [%s] has no matching '[%s]: <url>' link reference", v, v)
		}
	}

	// 3) web/site.json meta.content_version and meta.source_version (both carry
	// a leading "v"). source_version was outside the guard until v0.7.0, so a
	// one-field bump could go unnoticed.
	siteBytes, err := os.ReadFile(filepath.Join(repoRoot, "web", "site.json"))
	if err != nil {
		t.Fatalf("read web/site.json: %v", err)
	}
	var site struct {
		Meta struct {
			ContentVersion string `json:"content_version"`
			SourceVersion  string `json:"source_version"`
		} `json:"meta"`
	}
	if err := json.Unmarshal(siteBytes, &site); err != nil {
		t.Fatalf("parse web/site.json: %v", err)
	}
	contentVersion := strings.TrimPrefix(site.Meta.ContentVersion, "v")
	sourceVersion := strings.TrimPrefix(site.Meta.SourceVersion, "v")

	want := version
	if mcpxver.Version != want {
		t.Errorf("internal/version.Version = %q, want %q (ldflags default var)", mcpxver.Version, want)
	}
	if fileVersion != want {
		t.Errorf("VERSION file = %q, want %q (version var)", fileVersion, want)
	}
	if changelogHead != want {
		t.Errorf("CHANGELOG head = %q, want %q (version var)", changelogHead, want)
	}
	if contentVersion != want {
		t.Errorf("web/site.json meta.content_version = %q, want %q (version var)", site.Meta.ContentVersion, want)
	}
	if sourceVersion != want {
		t.Errorf("web/site.json meta.source_version = %q, want %q (version var)", site.Meta.SourceVersion, want)
	}
}

// changelogLinkRef matches a Keep-a-Changelog release link definition, e.g.
// "[0.7.0]: https://github.com/SuperMarioYL/mcpx/releases/tag/v0.7.0".
var changelogLinkRef = regexp.MustCompile(`^\[(\d+\.\d+\.\d+)\]:`)

// TestValidateClientIDs pins the --clients validation added so a typo'd id
// fails loudly instead of being silently dropped by DetectSubset.
func TestValidateClientIDs(t *testing.T) {
	if err := validateClientIDs(nil); err != nil {
		t.Fatalf("empty subset must pass, got %v", err)
	}
	if err := validateClientIDs([]string{"claude-code", "codex"}); err != nil {
		t.Fatalf("known ids must pass, got %v", err)
	}
	err := validateClientIDs([]string{"cursorr"})
	if err == nil {
		t.Fatal("unknown id must be rejected")
	}
	for _, want := range []string{"cursorr", "claude-code", "cursor"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q should mention %q (the typo and valid ids)", err, want)
		}
	}
}

// TestClientsFlagUnknownIDFailsFast runs the real command chain: the unknown
// id must produce an error before any detection/write output happens.
func TestClientsFlagUnknownIDFailsFast(t *testing.T) {
	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"add", "filesystem", "--clients", "cursorr", "--yes"})
	if err := root.Execute(); err == nil {
		t.Fatal("expected an error for an unknown --clients id")
	}
	if strings.Contains(out.String(), "Detected") || strings.Contains(out.String(), "backup") {
		t.Fatalf("command must fail before detection/write, got output:\n%s", out.String())
	}
}

// buildStub compiles the handshake testdata stub server once and returns its
// binary path. It supports MCPX_STUB_COUNT_FILE, which the stub appends a
// "spawn" line to at startup — used to count how many times the server was
// actually spawned.
func buildStub(t *testing.T) string {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go toolchain not on PATH; cannot build the stub server")
	}
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed to locate the test source")
	}
	repoRoot := filepath.Join(filepath.Dir(thisFile), "..", "..")
	bin := filepath.Join(t.TempDir(), "stubserver")
	cmd := exec.Command("go", "build", "-o", bin, "./internal/handshake/testdata/stubserver")
	cmd.Dir = repoRoot
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build stub server: %v\n%s", err, out)
	}
	return bin
}

// TestAddRunsHandshakeOnceForAllClients is the regression test for
// fix-handshake-once-per-add: the handshake is a pure function of the spec, so
// an add that writes into two clients must spawn the server exactly once.
func TestAddRunsHandshakeOnceForAllClients(t *testing.T) {
	stub := buildStub(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	if err := os.MkdirAll(filepath.Join(home, ".cursor"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{".claude.json", filepath.Join(".cursor", "mcp.json")} {
		if err := os.WriteFile(filepath.Join(home, name), []byte(`{}`), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	countFile := filepath.Join(t.TempDir(), "spawns")

	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{
		"add", "probe",
		"--command", stub,
		"--env", "MCPX_STUB_COUNT_FILE=" + countFile,
		"--yes",
	})
	if err := root.Execute(); err != nil {
		t.Fatalf("add failed: %v\n%s", err, out.String())
	}
	data, err := os.ReadFile(countFile)
	if err != nil {
		t.Fatalf("stub was never spawned (no count file):\n%s", out.String())
	}
	if n := strings.Count(string(data), "spawn"); n != 1 {
		t.Fatalf("handshake spawned the server %d times, want 1\noutput:\n%s", n, out.String())
	}
	// The two written clients each report the shared handshake verdict.
	if got := strings.Count(out.String(), "handshake OK"); got != 2 {
		t.Fatalf("want 2 per-client handshake lines, got %d\noutput:\n%s", got, out.String())
	}
	// Both configs actually received the entry.
	for _, name := range []string{".claude.json", filepath.Join(".cursor", "mcp.json")} {
		b, err := os.ReadFile(filepath.Join(home, name))
		if err != nil || !strings.Contains(string(b), `"probe"`) {
			t.Errorf("%s missing the written entry: %v %s", name, err, b)
		}
	}
}

// setupHome points every client adapter at a throwaway HOME and writes the given
// files (path relative to HOME -> content) before any command runs.
func setupHome(t *testing.T, files map[string]string) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	for name, content := range files {
		path := filepath.Join(home, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return home
}

// runRoot executes a fresh command tree and returns its combined output + error.
func runRoot(args ...string) (string, error) {
	var out bytes.Buffer
	root := newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs(args)
	err := root.Execute()
	return out.String(), err
}

// TestAddExitsNonZeroWhenHandshakeFails is the regression test for
// fix-add-exit-code-on-client-failure: a server that cannot be spawned used to
// print "! 0/1 clients ready" and still exit 0, so a script could not tell a
// total failure from a successful install.
func TestAddExitsNonZeroWhenHandshakeFails(t *testing.T) {
	home := setupHome(t, map[string]string{".claude.json": `{"mcpServers":{}}`})

	out, err := runRoot("add", "probe",
		"--command", filepath.Join(home, "missing-server"),
		"--clients", "claude-code", "--yes")
	if err == nil {
		t.Fatalf("a failed handshake must return an error (non-zero exit), output:\n%s", out)
	}
	if !strings.Contains(err.Error(), "1 of 1 client(s) failed") {
		t.Errorf("error should count the failed clients, got: %v", err)
	}
	if !strings.Contains(out, "handshake FAIL") || !strings.Contains(out, "0/1 clients ready") {
		t.Errorf("per-client verdict must still be printed, got:\n%s", out)
	}
}

// TestAddExitsNonZeroWhenClientWriteFails covers the other failure path: an
// unparseable client config leaves every write failed and must also exit non-zero.
func TestAddExitsNonZeroWhenClientWriteFails(t *testing.T) {
	setupHome(t, map[string]string{".claude.json": "not json at all"})

	out, err := runRoot("add", "probe", "--command", "npx", "--clients", "claude-code", "--yes")
	if err == nil {
		t.Fatalf("a failed write must return an error (non-zero exit), output:\n%s", out)
	}
	if !strings.Contains(out, "write FAIL") {
		t.Errorf("expected a write FAIL line, got:\n%s", out)
	}
	if !strings.Contains(err.Error(), "1 of 1 client(s) failed") {
		t.Errorf("error should count the failed clients, got: %v", err)
	}
}

// TestAddStaysSilentWhenNothingToDo pins the deliberate non-failures: no detected
// clients and dry-run keep exit 0.
func TestAddStaysSilentWhenNothingToDo(t *testing.T) {
	setupHome(t, nil)
	out, err := runRoot("add", "probe", "--command", "npx", "--yes")
	if err != nil {
		t.Fatalf("no detected clients is a no-op, not a failure: %v\n%s", err, out)
	}
}

// TestRemoveExitsNonZeroWhenClientFails is the regression test for
// fix-remove-hides-client-failures: the loop dropped errors and the closing line
// was an unconditional green "removed from 0 client(s)".
func TestRemoveExitsNonZeroWhenClientFails(t *testing.T) {
	setupHome(t, map[string]string{".claude.json": "not json at all"})

	out, err := runRoot("remove", "probe", "--clients", "claude-code", "--yes")
	if err == nil {
		t.Fatalf("a failed remove must return an error (non-zero exit), output:\n%s", out)
	}
	if !strings.Contains(out, "removed from 0 client(s), 1 failed") {
		t.Errorf("summary must report the failure, got:\n%s", out)
	}
}

// TestRemoveRejectsBlankServerName pins the guard that an empty name (which can
// never match an entry) fails before detection instead of reporting success.
func TestRemoveRejectsBlankServerName(t *testing.T) {
	setupHome(t, map[string]string{".claude.json": `{"mcpServers":{}}`})

	out, err := runRoot("remove", "", "--yes")
	if err == nil || !strings.Contains(err.Error(), "server name is required") {
		t.Fatalf("blank server name must be rejected, got err=%v output:\n%s", err, out)
	}
	if strings.Contains(out, "Detected") {
		t.Errorf("validation must happen before detection, got:\n%s", out)
	}
}

// TestAddNotesDroppedCodexComments pins the fix-codex-toml-rewrite-silent-comment-loss
// note: the TOML writer re-serializes the whole document, so it must say so when
// the config carried comments instead of dropping them silently.
func TestAddNotesDroppedCodexComments(t *testing.T) {
	stub := buildStub(t)
	setupHome(t, map[string]string{
		filepath.Join(".codex", "config.toml"): "# keep my notes\nmodel = \"gpt-5\"\n",
	})

	out, err := runRoot("add", "probe", "--command", stub, "--clients", "codex", "--yes")
	if err != nil {
		t.Fatalf("add failed: %v\n%s", err, out)
	}
	if !strings.Contains(out, "comments are not preserved") {
		t.Errorf("expected a comment-loss note in:\n%s", out)
	}
}
