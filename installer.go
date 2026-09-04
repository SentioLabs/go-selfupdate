package selfupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
)

// Installer replaces the running binary with the given release tag.
type Installer interface {
	Install(ctx context.Context, tag string) error
}

// safeTag is the only shape of tag ScriptInstaller will place on a command
// line: a release tag with no shell metacharacters.
var safeTag = regexp.MustCompile(`^v?[0-9A-Za-z][0-9A-Za-z.+-]*$`)

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

// Install runs the install script for tag.
func (s *ScriptInstaller) Install(ctx context.Context, tag string) error {
	line, err := s.command(tag)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, "bash", "-c", line)
	cmd.Stdout = s.Stdout
	if cmd.Stdout == nil {
		cmd.Stdout = os.Stdout
	}
	cmd.Stderr = s.Stderr
	if cmd.Stderr == nil {
		cmd.Stderr = os.Stderr
	}
	cmd.Stdin = s.Stdin
	if cmd.Stdin == nil {
		cmd.Stdin = os.Stdin
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("install script failed: %w", err)
	}
	return nil
}
