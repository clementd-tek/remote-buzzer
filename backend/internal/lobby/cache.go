package lobby

import "context"

// Cache is the interface a caching/persistence backend must satisfy for
// the Manager to keep the lobby directory in sync with it. It is
// implemented by internal/cache (Valkey/Redis) so this package stays
// storage-agnostic. A nil Cache is valid: the Manager simply runs
// in-memory only, which is fine for local development or a single
// instance without Valkey configured.
type Cache interface {
	Save(ctx context.Context, snapshot LobbySnapshot) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]LobbySnapshot, error)
}
