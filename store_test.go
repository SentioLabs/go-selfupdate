//nolint:testpackage // exercises MemStore and FuncStore alongside package internals
package selfupdate

import (
	"errors"
	"testing"
)

func TestMemStore(t *testing.T) {
	var s MemStore
	ch, err := s.Channel()
	if err != nil || ch != ChannelStable {
		t.Fatalf("empty MemStore must read as stable, got %q, %v", ch, err)
	}
	if err := s.SetChannel(ChannelRC); err != nil {
		t.Fatal(err)
	}
	ch, _ = s.Channel()
	if ch != ChannelRC {
		t.Fatalf("got %q", ch)
	}
}

func TestFuncStore(t *testing.T) {
	var saved Channel
	s := FuncStore{
		Get: func() (Channel, error) { return saved, nil },
		Set: func(c Channel) error { saved = c; return nil },
	}
	if err := s.SetChannel(ChannelNightly); err != nil {
		t.Fatal(err)
	}
	ch, err := s.Channel()
	if err != nil || ch != ChannelNightly {
		t.Fatalf("got %q, %v", ch, err)
	}

	boom := errors.New("boom")
	failing := FuncStore{
		Get: func() (Channel, error) { return "", boom },
		Set: func(Channel) error { return boom },
	}
	if _, err := failing.Channel(); !errors.Is(err, boom) {
		t.Fatal("Get error must propagate")
	}
	if err := failing.SetChannel(ChannelRC); !errors.Is(err, boom) {
		t.Fatal("Set error must propagate")
	}

	var nilFuncs FuncStore
	if _, err := nilFuncs.Channel(); err == nil {
		t.Fatal("nil Get must be an error, not a panic")
	}
	if err := nilFuncs.SetChannel(ChannelRC); err == nil {
		t.Fatal("nil Set must be an error, not a panic")
	}
}
