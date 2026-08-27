//go:build !windows

package handshake

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// TestTerminateProcReapsGrandchild guards the process-tree teardown fix. A
// runtime wrapper (the npx/uvx stand-in) forks a long-lived grandchild (the
// node/python stand-in). Killing only the wrapper — as the pre-v0.4.0 code did
// via cmd.Process.Kill() — orphans the grandchild to init; terminateProc must
// reap the whole process group so the grandchild dies with the wrapper.
func TestTerminateProcReapsGrandchild(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	pidFile := filepath.Join(t.TempDir(), "gc.pid")
	// The wrapper forks a 30s grandchild, then stays alive 10s as the "server".
	script := "sleep 30 & echo $! > " + pidFile + " ; sleep 10"
	cmd := exec.Command("bash", "-c", script)
	configureProc(cmd) // own process group: pgid == wrapper pid
	if err := cmd.Start(); err != nil {
		t.Fatalf("start wrapper: %v", err)
	}

	// Wait for the wrapper to record its grandchild's PID.
	gcPID := 0
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if b, err := os.ReadFile(pidFile); err == nil {
			if n, err := strconv.Atoi(strings.TrimSpace(string(b))); err == nil && n > 0 {
				gcPID = n
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	if gcPID == 0 {
		t.Fatal("wrapper never wrote its grandchild PID")
	}

	// Sanity: the grandchild is alive before teardown.
	if !processAlive(gcPID) {
		t.Fatal("grandchild died before teardown — test setup is wrong")
	}

	// Teardown the whole process group (wrapper + grandchild).
	terminateProc(cmd)
	_ = cmd.Wait() // reap the wrapper.

	// The grandchild must now be gone — reaped with the group, not orphaned.
	if !processGoneSoon(gcPID, 3*time.Second) {
		_ = exec.Command("kill", "-9", strconv.Itoa(gcPID)).Run()
		t.Fatalf("grandchild %d survived teardown — process tree leak", gcPID)
	}
}

func processAlive(pid int) bool {
	p, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return p.Signal(syscall.Signal(0)) == nil
}

func processGoneSoon(pid int, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !processAlive(pid) {
			return true
		}
		time.Sleep(30 * time.Millisecond)
	}
	return false
}
