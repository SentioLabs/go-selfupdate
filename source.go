package selfupdate

import "context"

// Release is the subset of a forge release the library needs.
type Release struct {
	Tag        string
	Prerelease bool
}

// Source lists releases. Implementations must return List newest first.
type Source interface {
	Latest(ctx context.Context) (Release, error)
	List(ctx context.Context, limit int) ([]Release, error)
}
