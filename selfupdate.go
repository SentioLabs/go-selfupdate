package selfupdate

import (
	"context"
	"errors"
	"io"
)

// ErrNotImplemented marks a stub that a later task replaces.
var ErrNotImplemented = errors.New("selfupdate: not implemented")

// Updater ties a CLI's identity to a Source, Store and Installer.
type Updater struct {
	Name      string        // binary name used in messages, e.g. "envctl"
	Version   string        // running version: "v1.2.3", "1.2.3" or "dev"
	Source    Source        // where releases are listed
	Store     Store         // where the channel is persisted; nil means stable only
	Installer Installer     // how the binary is replaced
	Channels  []ChannelSpec // nil means DefaultChannels
	Out       io.Writer     // nil means os.Stdout
	ErrOut    io.Writer     // nil means os.Stderr
	In        io.Reader     // nil means os.Stdin
	// PreInstall runs after the user confirms and before Installer.Install.
	PreInstall func(ctx context.Context, current, latest string) error
}

// UpdateOptions controls Update.
type UpdateOptions struct {
	Force bool // install even when up to date or newer
	Yes   bool // skip the confirmation prompt
	Check bool // print status only
}

// CheckResult is what Check found.
type CheckResult struct {
	Current string
	Latest  string
	Channel Channel
	Cmp     int // Compare(Current, Latest): 0 same, >0 update available, <0 current is newer
}

// CurrentChannel returns the persisted channel, or stable when none is set.
func (u *Updater) CurrentChannel() (Channel, error) { return "", ErrNotImplemented }

// Check resolves the latest release for the current channel without installing.
func (u *Updater) Check(ctx context.Context) (CheckResult, error) {
	return CheckResult{}, ErrNotImplemented
}

// Update checks and, unless opts.Check, installs after confirmation.
func (u *Updater) Update(ctx context.Context, opts UpdateOptions) error { return ErrNotImplemented }

// SwitchChannel validates, warns, and persists a new channel.
func (u *Updater) SwitchChannel(channel Channel, yes bool) error { return ErrNotImplemented }
