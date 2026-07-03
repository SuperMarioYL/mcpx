package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	toml "github.com/pelletier/go-toml/v2"
	"github.com/tidwall/gjson"

	"github.com/SuperMarioYL/mcpx/internal/clients"
)

// setHome points HOME at a temp dir so every adapter's ConfigPath resolves
// inside it, and returns the temp home.
func setHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	return home
}

func mustPath(t *testing.T, c clients.Client) string {
	t.Helper()
	p, ok, err := c.ConfigPath()
	if err != nil || !ok {
		t.Fatalf("config path for %s: %v", c.ID(), err)
	}
	return p
}

func writeFixture(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func backupsFor(t *testing.T, path string) []string {
	t.Helper()
	matches, err := filepath.Glob(path + ".mcpx.bak.*")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(matches)
	return matches
}

var fsSpec = clients.ServerSpec{
	Name:    "filesystem",
	Command: "npx",
	Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
	Env:     map[string]string{},
}

// ---- JSON adapter (claude-code, includes "type":"stdio") ------------------

func TestMergeJSON_ClaudeCode_PreservesSiblingsAndIsIdempotent(t *testing.T) {
	setHome(t)
	c, _ := clients.ByID("claude-code")
	path := mustPath(t, c)

	// Fixture mirrors the real ~/.claude.json shape: a big sibling "projects"
	// map, an unrelated top-level key, and a pre-existing MCP server.
	writeFixture(t, path, `{
  "numStartups": 42,
  "projects": {"/home/u/repo": {"allowedTools": ["Read"]}},
  "mcpServers": {
    "existing": {"type": "stdio", "command": "existing-bin", "args": [], "env": {}}
  }
}`)

	// First merge: changes the file, creates a backup.
	res, err := Merge(c, fsSpec, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Changed {
		t.Fatal("first merge should change the file")
	}
	if res.BackupPath == "" {
		t.Fatal("first merge should create a backup")
	}
	if _, err := os.Stat(res.BackupPath); err != nil {
		t.Fatalf("backup file missing: %v", err)
	}

	data, _ := os.ReadFile(path)

	// Unrelated top-level keys preserved.
	if gjson.GetBytes(data, "numStartups").Int() != 42 {
		t.Error("numStartups clobbered")
	}
	if !gjson.GetBytes(data, "projects./home/u/repo.allowedTools").Exists() {
		t.Error("projects map clobbered")
	}
	// Other mcpServers entry preserved.
	if gjson.GetBytes(data, "mcpServers.existing.command").String() != "existing-bin" {
		t.Error("existing MCP server clobbered")
	}
	// New entry written with the claude-code shape (type:stdio present).
	if got := gjson.GetBytes(data, "mcpServers.filesystem.type").String(); got != "stdio" {
		t.Errorf("claude-code entry should carry type:stdio, got %q", got)
	}
	if got := gjson.GetBytes(data, "mcpServers.filesystem.command").String(); got != "npx" {
		t.Errorf("new entry command = %q", got)
	}

	// Second identical merge: byte-stable no-op, no new backup, no change.
	before, _ := os.ReadFile(path)
	res2, err := Merge(c, fsSpec, false)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Changed {
		t.Error("second identical merge must be a no-op (idempotent)")
	}
	if res2.BackupPath != "" {
		t.Error("no-op merge must not create a backup")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Error("no-op merge must not rewrite the file")
	}
	if bs := backupsFor(t, path); len(bs) != 1 {
		t.Errorf("expected exactly 1 backup after idempotent re-merge, got %d", len(bs))
	}
}

func TestMergeJSON_ClaudeDesktop_OmitsTypeField(t *testing.T) {
	setHome(t)
	c, _ := clients.ByID("claude-desktop")
	path := mustPath(t, c)
	writeFixture(t, path, `{"coworkUserFilesPath":"/x","preferences":{"a":1},"mcpServers":{}}`)

	if _, err := Merge(c, fsSpec, false); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if gjson.GetBytes(data, "mcpServers.filesystem.type").Exists() {
		t.Error("claude-desktop entry must NOT include a type field")
	}
	if gjson.GetBytes(data, "coworkUserFilesPath").String() != "/x" {
		t.Error("claude-desktop sibling coworkUserFilesPath clobbered")
	}
	if gjson.GetBytes(data, "preferences.a").Int() != 1 {
		t.Error("claude-desktop sibling preferences clobbered")
	}
}

func TestMergeJSON_CreatesFileWhenAbsent(t *testing.T) {
	setHome(t)
	c, _ := clients.ByID("cursor")
	path := mustPath(t, c)

	res, err := Merge(c, fsSpec, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Created {
		t.Error("merge into absent config should report Created")
	}
	if res.BackupPath != "" {
		t.Error("no backup should be made for a freshly created file")
	}
	data, _ := os.ReadFile(path)
	if gjson.GetBytes(data, "mcpServers.filesystem.command").String() != "npx" {
		t.Error("entry not written to freshly created config")
	}
	var v any
	if err := json.Unmarshal(data, &v); err != nil {
		t.Errorf("created config is not valid JSON: %v", err)
	}
}

// ---- TOML adapter (codex) -------------------------------------------------

func TestMergeTOML_Codex_PreservesTablesAndIsIdempotent(t *testing.T) {
	setHome(t)
	c, _ := clients.ByID("codex")
	path := mustPath(t, c)

	// Fixture mirrors the real ~/.codex/config.toml: scalar top-level keys,
	// an unrelated [projects.*] table, and a pre-existing mcp_servers table
	// with a nested env table.
	writeFixture(t, path, `model = "gpt-5"
approval_policy = "never"

[features]
web_search = true

[projects."/home/u/repo"]
trust_level = "trusted"

[mcp_servers.existing]
command = "existing-bin"
args = ["serve"]

[mcp_servers.existing.env]
TOKEN = "keep-me"
`)

	specWithEnv := fsSpec
	specWithEnv.Env = map[string]string{"FOO": "bar"}

	res, err := Merge(c, specWithEnv, false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Changed {
		t.Fatal("first TOML merge should change the file")
	}
	if res.BackupPath == "" {
		t.Fatal("first TOML merge should create a backup")
	}

	root := map[string]any{}
	data, _ := os.ReadFile(path)
	if err := toml.Unmarshal(data, &root); err != nil {
		t.Fatalf("result is not valid TOML: %v\n%s", err, data)
	}

	// Unrelated scalar + tables preserved.
	if root["model"] != "gpt-5" {
		t.Error("top-level scalar 'model' clobbered")
	}
	features, _ := root["features"].(map[string]any)
	if features == nil || features["web_search"] != true {
		t.Error("[features] table clobbered")
	}
	projects, _ := root["projects"].(map[string]any)
	if projects == nil || projects["/home/u/repo"] == nil {
		t.Error("[projects.*] table clobbered")
	}

	servers, _ := root["mcp_servers"].(map[string]any)
	if servers == nil {
		t.Fatal("mcp_servers table missing")
	}
	// Pre-existing server + its nested env preserved.
	existing, _ := servers["existing"].(map[string]any)
	if existing == nil || existing["command"] != "existing-bin" {
		t.Error("existing mcp_server clobbered")
	}
	if env, _ := existing["env"].(map[string]any); env == nil || env["TOKEN"] != "keep-me" {
		t.Error("existing nested env table clobbered")
	}
	// New server written with its command + nested env.
	added, _ := servers["filesystem"].(map[string]any)
	if added == nil || added["command"] != "npx" {
		t.Fatalf("new codex mcp_server not written correctly: %+v", added)
	}
	if env, _ := added["env"].(map[string]any); env == nil || env["FOO"] != "bar" {
		t.Error("new server nested env not written")
	}

	// Idempotency: second identical merge is a no-op, no new backup.
	before, _ := os.ReadFile(path)
	res2, err := Merge(c, specWithEnv, false)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Changed {
		t.Error("second identical TOML merge must be a no-op")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Error("no-op TOML merge must not rewrite the file")
	}
	if bs := backupsFor(t, path); len(bs) != 1 {
		t.Errorf("expected exactly 1 backup after idempotent TOML re-merge, got %d", len(bs))
	}
}

// ---- remove + list --------------------------------------------------------

func TestRemoveJSON_PreservesOthers(t *testing.T) {
	setHome(t)
	c, _ := clients.ByID("claude-code")
	path := mustPath(t, c)
	writeFixture(t, path, `{"keepMe":true,"mcpServers":{"keep":{"command":"k","args":[],"env":{}}}}`)

	if _, err := Merge(c, fsSpec, false); err != nil {
		t.Fatal(err)
	}
	res, err := Remove(c, "filesystem", false)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Changed {
		t.Fatal("removing an existing entry should change the file")
	}
	if res.BackupPath == "" {
		t.Fatal("remove should create a backup")
	}
	data, _ := os.ReadFile(path)
	if gjson.GetBytes(data, "mcpServers.filesystem").Exists() {
		t.Error("removed entry still present")
	}
	if !gjson.GetBytes(data, "mcpServers.keep").Exists() {
		t.Error("remove clobbered an unrelated entry")
	}
	if gjson.GetBytes(data, "keepMe").Bool() != true {
		t.Error("remove clobbered an unrelated top-level key")
	}

	// Removing an absent entry is a no-op.
	res2, err := Remove(c, "filesystem", false)
	if err != nil {
		t.Fatal(err)
	}
	if res2.Changed {
		t.Error("removing an absent entry must be a no-op")
	}
}

func TestRemoveTOML_PreservesOthers(t *testing.T) {
	setHome(t)
	c, _ := clients.ByID("codex")
	path := mustPath(t, c)
	writeFixture(t, path, `model = "x"

[mcp_servers.keep]
command = "k"
args = []
`)
	if _, err := Merge(c, fsSpec, false); err != nil {
		t.Fatal(err)
	}
	if _, err := Remove(c, "filesystem", false); err != nil {
		t.Fatal(err)
	}
	root := map[string]any{}
	data, _ := os.ReadFile(path)
	if err := toml.Unmarshal(data, &root); err != nil {
		t.Fatal(err)
	}
	if root["model"] != "x" {
		t.Error("remove clobbered top-level scalar")
	}
	servers, _ := root["mcp_servers"].(map[string]any)
	if servers == nil || servers["keep"] == nil {
		t.Error("remove clobbered an unrelated mcp_server")
	}
	if servers["filesystem"] != nil {
		t.Error("removed codex entry still present")
	}
}

func TestList(t *testing.T) {
	setHome(t)
	cc, _ := clients.ByID("claude-code")
	if _, err := Merge(cc, fsSpec, false); err != nil {
		t.Fatal(err)
	}
	got, err := List(cc)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Name != "filesystem" || got[0].Command != "npx" {
		t.Fatalf("list mismatch: %+v", got)
	}
	if strings.Join(got[0].Args, " ") != "-y @modelcontextprotocol/server-filesystem ." {
		t.Errorf("list args mismatch: %v", got[0].Args)
	}
}

func TestDryRunWritesNothing(t *testing.T) {
	setHome(t)
	c, _ := clients.ByID("claude-code")
	path := mustPath(t, c)
	writeFixture(t, path, `{"mcpServers":{}}`)
	before, _ := os.ReadFile(path)

	res, err := Merge(c, fsSpec, true)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Changed {
		t.Error("dry-run should report that a change WOULD happen")
	}
	if res.BackupPath != "" {
		t.Error("dry-run must not create a backup")
	}
	after, _ := os.ReadFile(path)
	if string(before) != string(after) {
		t.Error("dry-run must not modify the file")
	}
	if bs := backupsFor(t, path); len(bs) != 0 {
		t.Errorf("dry-run must not create backups, found %d", len(bs))
	}
}
