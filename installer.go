package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"
)

// Installer replaces the running binary with the given release tag.
type Installer interface {
	Install(ctx context.Context, tag string) error
}

// safeTag is the only shape of tag ScriptInstaller will place on a command
// line: a release tag with no shell metacharacters.
var safeTag = regexp.MustCompile(`^v?[0-9A-Za-z][0-9A-Za-z.+-]*$`)

// pipelineWaitDelay bounds how long Install waits for the process group to
// exit after a SIGTERM before Go escalates to SIGKILL on the direct child.
const pipelineWaitDelay = 5 * time.Second

// ScriptInstaller runs `curl -fsSL <ScriptURL> | bash -s -- --force --tag=<tag>`
// with the process's stdio attached. Both arc and envctl install this way.
type ScriptInstaller struct {
	ScriptURL string
	Stdout    io.Writer // default os.Stdout
	Stderr    io.Writer // default os.Stderr
	Stdin     io.Reader // default os.Stdin
}

// command builds the shell line, refusing tags that could inject commands.
func (s *ScriptInstaller) command(tag string) (string, error) {
	if s.ScriptURL == "" {
		return "", errors.New("selfupdate: ScriptInstaller.ScriptURL is empty")
	}
	if !safeTag.MatchString(tag) {
		return "", fmt.Errorf("selfupdate: refusing to install unsafe tag %q", tag)
	}
	return "curl -fsSL " + s.ScriptURL + " | bash -s -- --force --tag=" + tag, nil
}

// stdinIsTerminal reports whether r is the controlling terminal. It returns
// true only for an *os.File whose mode has os.ModeCharDevice set. A nil
// reader, a non-file reader, or a failed Stat all return false.
func stdinIsTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok || f == nil {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// Install runs the install script for tag. It runs under pipefail so a
// curl failure fails the pipeline instead of exiting 0 on empty stdin.
// With a terminal on stdin, the pipeline shares the CLI's process group.
// That keeps prompts such as sudo working and lets Ctrl-C reach every
// process. Without a terminal, cancelling ctx stops the whole pipeline.
func (s *ScriptInstaller) Install(ctx context.Context, tag string) error {
	line, err := s.command(tag)
	if err != nil {
		return err
	}
	stdin := s.Stdin
	if stdin == nil {
		stdin = os.Stdin
	}
	cmd := exec.CommandContext(ctx, "bash", "-o", "pipefail", "-c", line)
	if !stdinIsTerminal(stdin) {
		setProcessGroup(cmd)
	}
	cmd.WaitDelay = pipelineWaitDelay
	cmd.Stdout = s.Stdout
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = s.Stderr
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = stdin
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install script failed: %w", err)
	}
	return nil
}
