package selfupdate

// Store persists the selected channel. Consumers adapt their own config.
type Store interface {
	Channel() (Channel, error)
	SetChannel(Channel) error
}
