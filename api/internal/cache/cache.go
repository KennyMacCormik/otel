package cache

import (
	"github.com/KennyMacCormik/otel/backend/pkg/cache"
	"github.com/KennyMacCormik/otel/backend/pkg/cache/impl/sync_map"
	"github.com/KennyMacCormik/otel/backend/pkg/cache/wrappers/sharded_cache"
	"github.com/KennyMacCormik/otel/backend/pkg/cache/wrappers/ttl_cache"
)

func NewCache() (cache.CacheInterface, error) {
	fn := func() cache.CacheInterface {
		c, _ := ttl_cache.NewTtlCache(sync_map.NewSyncMapCache())
		return c
	}
	c, err := sharded_cache.NewShardedCache(fn)
	if err != nil {
		return nil, err
	}
	return c, nil
}
