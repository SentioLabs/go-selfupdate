package selfupdate

import "errors"

// Store persists the selected channel. Consumers adapt their own config.
type Store interface {
	Channel() (Channel, error)
	SetChannel(Channel) error
}

// ErrNoStore is returned by a FuncStore whose closures are nil.
var ErrNoStore = errors.New("selfupdate: channel store not configured")

// FuncStore adapts two closures to Store.
type FuncStore struct {
	Get func() (Channel, error)
	Set func(Channel) error
}

// Channel calls Get. A nil Get is ErrNoStore.
func (f FuncStore) Channel() (Channel, error) {
	if f.Get == nil {
		return "", ErrNoStore
	}
	return f.Get()
}

// SetChannel calls Set. A nil Set is ErrNoStore.
func (f FuncStore) SetChannel(c Channel) error {
	if f.Set == nil {
		return ErrNoStore
	}
	return f.Set(c)
}

// MemStore is an in-memory Store for tests and for CLIs without persistence.
// An empty Current reads as stable.
type MemStore struct{ Current Channel }

// Channel returns Current, or stable when unset.
func (m *MemStore) Channel() (Channel, error) {
	if m.Current == "" {
		return ChannelStable, nil
	}
	return m.Current, nil
}

// SetChannel records c.
func (m *MemStore) SetChannel(c Channel) error {
	m.Current = c
	return nil
}
