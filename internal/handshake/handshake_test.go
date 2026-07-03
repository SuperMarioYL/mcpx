package handshake

import (
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/SuperMarioYL/mcpx/internal/clients"
)

// buildStub compiles the testdata stub server once and returns its binary path.
func buildStub(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	bin := filepath.Join(dir, "stubserver")
	cmd := exec.Command("go", "build", "-o", bin, "./testdata/stubserver")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build stub server: %v\n%s", err, out)
	}
	return bin
}

func specFor(bin, mode string) clients.ServerSpec {
	return clients.ServerSpec{
		Name:    "stub",
		Command: bin,
		Env:     map[string]string{"MCPX_STUB_MODE": mode},
	}
}

func TestHandshakeOK(t *testing.T) {
	bin := buildStub(t)
	res := TestContext(t.Context(), specFor(bin, "ok"), 5*time.Second)
	if res.Status != StatusOK {
		t.Fatalf("want OK, got %s (reason=%q)", res.Status, res.Reason)
	}
	if res.ToolCount != 2 {
		t.Fatalf("want 2 tools, got %d", res.ToolCount)
	}
}

func TestHandshakeToolsError(t *testing.T) {
	bin := buildStub(t)
	res := TestContext(t.Context(), specFor(bin, "error"), 5*time.Second)
	if res.Status != StatusFail {
		t.Fatalf("want FAIL on tools/list error, got %s (reason=%q)", res.Status, res.Reason)
	}
	if res.Reason == "" {
		t.Fatalf("expected a failure reason")
	}
}

func TestHandshakeCrash(t *testing.T) {
	bin := buildStub(t)
	res := TestContext(t.Context(), specFor(bin, "crash"), 5*time.Second)
	if res.Status != StatusFail {
		t.Fatalf("want FAIL when server crashes, got %s", res.Status)
	}
}

func TestHandshakeSilentTimesOut(t *testing.T) {
	bin := buildStub(t)
	start := time.Now()
	res := TestContext(t.Context(), specFor(bin, "silent"), 600*time.Millisecond)
	elapsed := time.Since(start)
	if res.Status != StatusWarn {
		t.Fatalf("want WARN (timeout, fail-soft), got %s (reason=%q)", res.Status, res.Reason)
	}
	if elapsed > 3*time.Second {
		t.Fatalf("handshake hung for %v; timeout not enforced", elapsed)
	}
}

func TestHandshakeMissingCommand(t *testing.T) {
	res := Test(clients.ServerSpec{Name: "nope", Command: ""})
	if res.Status != StatusFail {
		t.Fatalf("want FAIL for empty command, got %s", res.Status)
	}
}

func TestHandshakeUnknownBinary(t *testing.T) {
	res := TestContext(t.Context(), clients.ServerSpec{
		Name:    "nope",
		Command: filepath.Join(t.TempDir(), "does-not-exist-binary"),
	}, 2*time.Second)
	if res.Status != StatusFail {
		t.Fatalf("want FAIL for nonexistent binary, got %s", res.Status)
	}
}
