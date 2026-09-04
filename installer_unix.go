//go:build unix

package selfupdate

import (
	"os"
	"os/exec"
	"syscall"
)

// setProcessGroup puts cmd in its own process group so Install can signal
// the whole pipeline, not just the direct child, when its context is
// cancelled.
func setProcessGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		if err == syscall.ESRCH {
			// The group already exited, so treat a cancel racing a
			// normal exit as done rather than as a spurious error.
			return os.ErrProcessDone
		}
		return err
	}
}
