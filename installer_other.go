//go:build !unix

package selfupdate

import "os/exec"

// setProcessGroup is a no-op on platforms without POSIX process groups.
func setProcessGroup(_ *exec.Cmd) {}
