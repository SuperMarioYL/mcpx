package clients

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// fakeHome points HOME (and APPDATA on Windows) at a temp dir so ConfigPath
// resolves inside it, then returns the resolved config path for each client.
func fakeHome(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows os.UserHomeDir
	t.Setenv("APPDATA", filepath.Join(home, "AppData", "Roaming"))
	return home
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func pathFor(t *testing.T, id string) string {
	t.Helper()
	c, ok := ByID(id)
	if !ok {
		t.Fatalf("unknown client id %q", id)
	}
	p, ok, err := c.ConfigPath()
	if err != nil || !ok {
		t.Fatalf("resolve config path for %s: ok=%v err=%v", id, ok, err)
	}
	return p
}

func detectedByID(list []DetectedClient, id string) (DetectedClient, bool) {
	for _, dc := range list {
		if dc.Adapter.ID() == id {
			return dc, true
		}
	}
	return DetectedClient{}, false
}

func TestDetectPresentAndAbsent(t *testing.T) {
	fakeHome(t)

	// Present: claude-code + codex. Absent: claude-desktop + cursor.
	writeFile(t, pathFor(t, "claude-code"), `{"mcpServers":{}}`)
	writeFile(t, pathFor(t, "codex"), "model = \"x\"\n")

	got := Detect()

	if dc, ok := detectedByID(got, "claude-code"); !ok || !dc.Exists || dc.Skipped {
		t.Errorf("claude-code should be detected: %+v", dc)
	}
	if dc, ok := detectedByID(got, "codex"); !ok || !dc.Exists || dc.Skipped {
		t.Errorf("codex should be detected: %+v", dc)
	}
	if dc, ok := detectedByID(got, "cursor"); !ok || dc.Exists {
		t.Errorf("cursor should NOT be detected (no config): %+v", dc)
	}
	if dc, ok := detectedByID(got, "claude-desktop"); !ok || dc.Exists {
		t.Errorf("claude-desktop should NOT be detected (no config): %+v", dc)
	}

	present := Present(got)
	if len(present) != 2 {
		t.Fatalf("want 2 present clients, got %d: %+v", len(present), present)
	}
}

func TestDetectDirectoryIsNotAConfig(t *testing.T) {
	fakeHome(t)
	// Create a directory where the config file would be — must not count.
	p := pathFor(t, "cursor")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	got := Detect()
	if dc, ok := detectedByID(got, "cursor"); !ok || dc.Exists {
		t.Errorf("a directory at the config path must not count as detected: %+v", dc)
	}
}

func TestDetectSubset(t *testing.T) {
	fakeHome(t)
	writeFile(t, pathFor(t, "claude-code"), `{"mcpServers":{}}`)
	writeFile(t, pathFor(t, "codex"), "model = \"x\"\n")

	got := DetectSubset([]string{"claude-code"})
	if len(got) != 1 {
		t.Fatalf("subset should return exactly 1 adapter, got %d", len(got))
	}
	if got[0].Adapter.ID() != "claude-code" {
		t.Fatalf("subset returned wrong adapter: %s", got[0].Adapter.ID())
	}

	// Unknown ids are ignored; empty subset falls back to all.
	if len(DetectSubset(nil)) != len(All()) {
		t.Fatalf("nil subset should detect all clients")
	}
	if got := DetectSubset([]string{"nope"}); len(got) != 0 {
		t.Fatalf("unknown-only subset should be empty, got %d", len(got))
	}
}

func TestConfigPathsResolveOnThisOS(t *testing.T) {
	fakeHome(t)
	for _, c := range All() {
		p, ok, err := c.ConfigPath()
		if err != nil || !ok {
			t.Errorf("%s ConfigPath unresolved on %s: ok=%v err=%v", c.ID(), runtime.GOOS, ok, err)
			continue
		}
		if !filepath.IsAbs(p) {
			t.Errorf("%s ConfigPath not absolute: %q", c.ID(), p)
		}
	}
}
