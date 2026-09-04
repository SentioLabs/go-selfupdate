// Package selfupdate implements self-update with release channels for Go
// CLIs that publish GitHub releases. A consumer wires an Updater with a
// Source (where releases come from), a Store (where the chosen channel is
// persisted) and an Installer (how the binary is replaced).
package selfupdate

import (
	"context"
	"fmt"
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

// Resolve returns the tag a user on channel should be offered. An empty
// channel means stable, which uses the forge's latest release. A patterned
// channel returns its newest matching tag unless the newest stable release
// is newer, in which case the stable tag is returned. A patterned channel
// with no match falls back to the newest stable release. An unknown channel
// is an error, as is a channel with no release at all.
func Resolve(ctx context.Context, src Source, channel Channel, specs []ChannelSpec) (string, error) {
	if specs == nil {
		specs = DefaultChannels
	}
	if channel == "" {
		channel = ChannelStable
	}
	spec, ok := LookupChannel(specs, channel)
	if !ok {
		return "", fmt.Errorf("unknown channel %q", channel)
	}
	if spec.Pattern == nil {
		rel, err := src.Latest(ctx)
		if err != nil {
			return "", err
		}
		return rel.Tag, nil
	}

	releases, err := src.List(ctx, maxPerPage)
	if err != nil {
		return "", err
	}

	channelTag, stableTag := newestMatch(releases, spec.Pattern)
	switch {
	case channelTag != "" && stableTag != "":
		if semver.Compare(stableTag, channelTag) > 0 {
			return stableTag, nil
		}
		return channelTag, nil
	case channelTag != "":
		return channelTag, nil
	case stableTag != "":
		return stableTag, nil
	default:
		return "", fmt.Errorf("no %s release found", channel)
	}
}

// newestMatch scans releases, which List returns newest first, and returns
// the first tag matching pattern and the first non-prerelease tag.
func newestMatch(releases []Release, pattern *regexp.Regexp) (channelTag, stableTag string) {
	for _, r := range releases {
		if channelTag == "" && pattern.MatchString(r.Tag) {
			channelTag = r.Tag
		}
		if stableTag == "" && !r.Prerelease {
			stableTag = r.Tag
		}
		if channelTag != "" && stableTag != "" {
			break
		}
	}
	return channelTag, stableTag
}
