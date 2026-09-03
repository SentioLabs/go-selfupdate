// Package selfupdate's own test package is used (not selfupdate_test) so the
// contract assertions below read exactly like consumer code: the explicit
// type on each "var _ T = value" line is the assertion, not incidental style,
// so it is kept even where staticcheck would suggest inferring it.
//
//nolint:staticcheck,testpackage // contract assertions kept verbatim per plan; see design-m1.md
package selfupdate

import (
	"regexp"
	"testing"
)

// --- Contract assertions ---
// These verify the design contracts. Do NOT modify without updating the plan.

func TestContracts(t *testing.T) {
	var _ Channel = ChannelStable
	var _ Channel = ChannelRC
	var _ Channel = ChannelNightly

	spec := ChannelSpec{}
	var _ Channel = spec.Name
	var _ *regexp.Regexp = spec.Pattern
	var _ string = spec.Warning

	r := Release{}
	var _ string = r.Tag
	var _ bool = r.Prerelease

	u := Updater{}
	var _ string = u.Name
	var _ string = u.Version
	var _ Source = u.Source
	var _ Store = u.Store
	var _ Installer = u.Installer
	var _ []ChannelSpec = u.Channels

	o := UpdateOptions{}
	var _ bool = o.Force
	var _ bool = o.Yes
	var _ bool = o.Check

	c := CheckResult{}
	var _ int = c.Cmp
	var _ Channel = c.Channel

	if len(DefaultChannels) != 3 {
		t.Fatalf("DefaultChannels: want 3, got %d", len(DefaultChannels))
	}
	if DefaultChannels[0].Pattern != nil {
		t.Fatal("stable must have no pattern")
	}
	if _, ok := LookupChannel(DefaultChannels, ChannelRC); !ok {
		t.Fatal("rc must be a default channel")
	}
	if _, ok := LookupChannel(DefaultChannels, "beta"); ok {
		t.Fatal("beta must not be a default channel")
	}
}

func TestDefaultPatterns(t *testing.T) {
	rc, _ := LookupChannel(DefaultChannels, ChannelRC)
	nightly, _ := LookupChannel(DefaultChannels, ChannelNightly)
	for _, tag := range []string{"v1.2.3-rc.1", "v1.2.3-rc1", "v0.16.0-rc9"} {
		if !rc.Pattern.MatchString(tag) {
			t.Errorf("rc pattern must match %s", tag)
		}
	}
	for _, tag := range []string{"v1.2.3", "v1.2.3-nightly.20260903", "v1.2.3-rc"} {
		if rc.Pattern.MatchString(tag) {
			t.Errorf("rc pattern must not match %s", tag)
		}
	}
	if !nightly.Pattern.MatchString("v0.6.1-nightly.20260904") {
		t.Error("nightly pattern must match a dated tag")
	}
	if nightly.Pattern.MatchString("v0.6.1-nightly.2026") {
		t.Error("nightly pattern requires an 8-digit date")
	}
}
