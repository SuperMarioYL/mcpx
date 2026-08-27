//go:build !windows

package handshake

import (
	"os/exec"
	"syscall"
)

// configureProc puts the server in its own process group so the whole runtime
// tree can be torn down together. npx/uvx spawn a grandchild (node/python); if
// the wrapper is the only thing in a group, killing the group reaps the
// grandchild too instead of orphaning it to init.
func configureProc(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// terminateProc kills the server's whole process group. With Setpgid the
// spawned process is the group leader (pgid == pid), so -pid targets the group
// (the wrapper plus any runtime grandchild it forked). ESRCH (the group is
// already gone) is ignored. This replaces the old single-PID SIGKILL that left
// the npx/uvx grandchild orphaned and running.
func terminateProc(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
}
