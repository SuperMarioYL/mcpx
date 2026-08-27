//go:build windows

package handshake

import (
	"fmt"
	"os/exec"
	"syscall"
)

// configureProc puts the server in its own process group so the whole runtime
// tree can be torn down together. npx/uvx spawn a grandchild (node/python); a
// dedicated group lets terminateProc reap the whole tree instead of orphaning
// the grandchild when the wrapper dies.
func configureProc(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// terminateProc kills the server's whole process tree. Windows has no
// negative-pgid kill, so taskkill /T walks and terminates the descendant tree
// rooted at the wrapper (the npx/uvx grandchild included). /F forces termination.
// A missing PID is a no-op; taskkill's own non-zero exit (already gone) is
// ignored so repeated teardown stays idempotent.
func terminateProc(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = exec.Command("taskkill", "/PID", fmt.Sprint(cmd.Process.Pid), "/T", "/F").Run()
}
