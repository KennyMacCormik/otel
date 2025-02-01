package storage

import (
	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	"github.com/KennyMacCormik/otel/backend/pkg/cache/impl/sync_map"
)

func NewStorage() cache.CacheInterface {
	return sync_map.NewSyncMapCache()
}
