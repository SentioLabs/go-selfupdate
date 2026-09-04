//nolint:testpackage // needs the unexported ScriptInstaller.command method
package selfupdate

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// requireCurlAndBash skips the test when either tool is missing from PATH.
func requireCurlAndBash(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("curl"); err != nil {
		t.Skip("curl not on PATH")
	}
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not on PATH")
	}
}

func TestScriptInstaller_Command(t *testing.T) {
	s := &ScriptInstaller{ScriptURL: "https://example.com/install.sh"}
	got, err := s.command("v1.2.3-rc.1")
	if err != nil {
		t.Fatal(err)
	}
	want := "curl -fsSL https://example.com/install.sh | bash -s -- --force --tag=v1.2.3-rc.1"
	if got != want {
		t.Fatalf("got %q", got)
	}
}

func TestScriptInstaller_RejectsUnsafeTag(t *testing.T) {
	s := &ScriptInstaller{ScriptURL: "https://example.com/install.sh"}
	for _, bad := range []string{"", "v1; rm -rf /", "v1 $(x)", "v1`x`", "v1&&x", "v1|x", "v1 x", "../x"} {
		if _, err := s.command(bad); err == nil {
			t.Errorf("tag %q must be rejected", bad)
		}
		if err := s.Install(context.Background(), bad); err == nil {
			t.Errorf("Install with tag %q must fail before running anything", bad)
		}
	}
}

func TestScriptInstaller_RunsScriptViaFileURL(t *testing.T) {
	requireCurlAndBash(t)
	// curl accepts file:// URLs, so a local script stands in for the remote one.
	dir := t.TempDir()
	script := filepath.Join(dir, "install.sh")
	body := "#!/usr/bin/env bash\necho \"args: $*\"\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil { //nolint:gosec // script must be executable
		t.Fatal(err)
	}
	var out bytes.Buffer
	s := &ScriptInstaller{ScriptURL: "file://" + script, Stdout: &out, Stderr: &out}
	if err := s.Install(context.Background(), "v9.9.9"); err != nil {
		t.Fatalf("install: %v, output: %s", err, out.String())
	}
	if !strings.Contains(out.String(), "args: --force --tag=v9.9.9") {
		t.Fatalf("script did not receive the expected arguments: %s", out.String())
	}
}

func TestScriptInstaller_NonZeroExitIsError(t *testing.T) {
	requireCurlAndBash(t)
	dir := t.TempDir()
	script := filepath.Join(dir, "install.sh")
	body := "#!/usr/bin/env bash\nexit 3\n"
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil { //nolint:gosec // script must be executable
		t.Fatal(err)
	}
	var out bytes.Buffer
	s := &ScriptInstaller{ScriptURL: "file://" + script, Stdout: &out, Stderr: &out}
	if err := s.Install(context.Background(), "v1.0.0"); err == nil {
		t.Fatal("expected error from exit 3")
	}
}
