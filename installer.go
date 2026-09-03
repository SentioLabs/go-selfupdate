package selfupdate

import "context"

// Installer replaces the running binary with the given release tag.
type Installer interface {
	Install(ctx context.Context, tag string) error
}
