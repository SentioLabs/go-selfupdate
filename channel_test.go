// Package selfupdate's own test package is used (not selfupdate_test) per
// the task spec, so these tests can be extended later to exercise unexported
// helpers without a second test package.
//
//nolint:testpackage // internal package per task spec envctl-04qa.00bow5.1.1.4
package selfupdate

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// tagV0110RC3 and tagV0100 name literals repeated across the fixtures below
// so their reuse doesn't trip goconst.
const (
	tagV0110RC3 = "v0.11.0-rc3"
	tagV0100    = "v0.10.0"
)

// fakeSource serves canned releases. latestErr and listErr force failures.
type fakeSource struct {
	latest    Release
	list      []Release
	latestErr error
	listErr   error
	listCalls int
}

func (f *fakeSource) Latest(context.Context) (Release, error) { return f.latest, f.latestErr }
func (f *fakeSource) List(context.Context, int) ([]Release, error) {
	f.listCalls++
	return f.list, f.listErr
}

// fixtureReleases mirrors arc's fixture: undotted and dotted rc tags,
// nightlies of the next version, one stable.
func fixtureReleases() []Release {
	return []Release{
		{Tag: tagV0110RC3, Prerelease: true},
		{Tag: "v0.11.0-rc2", Prerelease: true},
		{Tag: "v0.11.0-rc.1", Prerelease: true},
		{Tag: "v0.11.0-nightly.20260302", Prerelease: true},
		{Tag: "v0.11.0-nightly.20260301", Prerelease: true},
		{Tag: tagV0100},
	}
}

func TestResolve_StableUsesLatest(t *testing.T) {
	src := &fakeSource{latest: Release{Tag: tagV0100}, list: fixtureReleases()}
	for _, ch := range []Channel{ChannelStable, ""} {
		tag, err := Resolve(context.Background(), src, ch, DefaultChannels)
		if err != nil || tag != tagV0100 {
			t.Fatalf("channel %q: got %q, %v", ch, tag, err)
		}
	}
	if src.listCalls != 0 {
		t.Fatal("stable must not list releases")
	}
}

func TestResolve_RC(t *testing.T) {
	src := &fakeSource{list: fixtureReleases()}
	tag, err := Resolve(context.Background(), src, ChannelRC, DefaultChannels)
	if err != nil || tag != tagV0110RC3 {
		t.Fatalf("got %q, %v", tag, err)
	}
}

func TestResolve_RCStableIsNewer(t *testing.T) {
	src := &fakeSource{list: []Release{
		{Tag: "v0.12.0"},
		{Tag: tagV0110RC3, Prerelease: true},
		{Tag: tagV0100},
	}}
	tag, err := Resolve(context.Background(), src, ChannelRC, DefaultChannels)
	if err != nil || tag != "v0.12.0" {
		t.Fatalf("got %q, %v", tag, err)
	}
}

func TestResolve_Nightly(t *testing.T) {
	src := &fakeSource{list: fixtureReleases()}
	tag, err := Resolve(context.Background(), src, ChannelNightly, DefaultChannels)
	if err != nil || tag != "v0.11.0-nightly.20260302" {
		t.Fatalf("got %q, %v", tag, err)
	}
}

func TestResolve_NightlyStableIsNewer(t *testing.T) {
	src := &fakeSource{list: []Release{
		{Tag: "v0.11.0"},
		{Tag: "v0.10.0-nightly.20260302", Prerelease: true},
	}}
	tag, err := Resolve(context.Background(), src, ChannelNightly, DefaultChannels)
	if err != nil || tag != "v0.11.0" {
		t.Fatalf("got %q, %v", tag, err)
	}
}

func TestResolve_NoChannelMatchFallsBackToStable(t *testing.T) {
	src := &fakeSource{list: []Release{{Tag: tagV0100}}}
	tag, err := Resolve(context.Background(), src, ChannelNightly, DefaultChannels)
	if err != nil || tag != tagV0100 {
		t.Fatalf("got %q, %v", tag, err)
	}
}

func TestResolve_NothingFound(t *testing.T) {
	src := &fakeSource{}
	_, err := Resolve(context.Background(), src, ChannelRC, DefaultChannels)
	if err == nil || !strings.Contains(err.Error(), "no rc release found") {
		t.Fatalf("got %v", err)
	}
}

func TestResolve_UnknownChannel(t *testing.T) {
	_, err := Resolve(context.Background(), &fakeSource{}, "beta", DefaultChannels)
	if err == nil || !strings.Contains(err.Error(), "unknown channel") {
		t.Fatalf("got %v", err)
	}
}

func TestResolve_NilSpecsMeansDefaults(t *testing.T) {
	src := &fakeSource{list: fixtureReleases()}
	tag, err := Resolve(context.Background(), src, ChannelRC, nil)
	if err != nil || tag != tagV0110RC3 {
		t.Fatalf("got %q, %v", tag, err)
	}
}

func TestResolve_SourceErrorsPropagate(t *testing.T) {
	boom := errors.New("boom")
	stableSrc := &fakeSource{latestErr: boom}
	if _, err := Resolve(context.Background(), stableSrc, ChannelStable, nil); !errors.Is(err, boom) {
		t.Fatalf("stable: %v", err)
	}
	if _, err := Resolve(context.Background(), &fakeSource{listErr: boom}, ChannelRC, nil); !errors.Is(err, boom) {
		t.Fatalf("rc: %v", err)
	}
}
