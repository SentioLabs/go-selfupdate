// Package selfupdate implements self-update with release channels for Go
// CLIs that publish GitHub releases. A consumer wires an Updater with a
// Source (where releases come from), a Store (where the chosen channel is
// persisted) and an Installer (how the binary is replaced).
package selfupdate

import (
	"regexp"

	"golang.org/x/mod/semver"
)

// Channel names a release stream.
type Channel string

// Built-in channel names.
const (
	ChannelStable  Channel = "stable"
	ChannelRC      Channel = "rc"
	ChannelNightly Channel = "nightly"
)

// ChannelSpec describes one channel. Pattern is nil for stable, which uses
// the forge's latest release instead of tag matching. Warning is printed
// before a switch to the channel; empty means no confirmation prompt.
type ChannelSpec struct {
	Name    Channel
	Pattern *regexp.Regexp
	Warning string
}

// DefaultChannels are stable, rc (v1.2.3-rc.1 or v1.2.3-rc1) and nightly
// (v1.2.3-nightly.20260903).
var DefaultChannels = []ChannelSpec{
	{Name: ChannelStable},
	{
		Name:    ChannelRC,
		Pattern: regexp.MustCompile(`^v\d+\.\d+\.\d+-rc\.?\d+$`),
		Warning: "Release candidates may contain bugs that have not been fully tested.",
	},
	{
		Name:    ChannelNightly,
		Pattern: regexp.MustCompile(`^v\d+\.\d+\.\d+-nightly\.\d{8}$`),
		Warning: "Nightly builds come from the latest main branch and may be unstable.",
	},
}

// LookupChannel returns the spec named by channel, or false.
func LookupChannel(specs []ChannelSpec, channel Channel) (ChannelSpec, bool) {
	for _, s := range specs {
		if s.Name == channel {
			return s, true
		}
	}
	return ChannelSpec{}, false
}

// isValidSemver reports whether tag parses as semver with a leading v.
// It exists so channel.go references semver now; Resolve uses it later.
//
//nolint:unused // referenced by Resolve in a later task
func isValidSemver(tag string) bool { return semver.IsValid(tag) }
